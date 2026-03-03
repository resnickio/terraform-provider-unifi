package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &SettingRadiusResource{}
	_ resource.ResourceWithImportState = &SettingRadiusResource{}
)

type SettingRadiusResource struct {
	client *AutoLoginClient
}

type SettingRadiusResourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	SiteID                types.String   `tfsdk:"site_id"`
	Enabled               types.Bool     `tfsdk:"enabled"`
	AccountingEnabled     types.Bool     `tfsdk:"accounting_enabled"`
	AuthPort              types.Int64    `tfsdk:"auth_port"`
	AcctPort              types.Int64    `tfsdk:"acct_port"`
	XSecret               types.String   `tfsdk:"x_secret"`
	TunneledReply         types.Bool     `tfsdk:"tunneled_reply"`
	InterimUpdateInterval types.Int64    `tfsdk:"interim_update_interval"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

func NewSettingRadiusResource() resource.Resource {
	return &SettingRadiusResource{}
}

func (r *SettingRadiusResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_radius"
}

func (r *SettingRadiusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi site RADIUS server settings. Singleton resource — one per site. Delete resets to defaults.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the RADIUS server.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"accounting_enabled": schema.BoolAttribute{
				Description: "Enable RADIUS accounting.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"auth_port": schema.Int64Attribute{
				Description: "RADIUS authentication port. Defaults to 1812.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1812),
			},
			"acct_port": schema.Int64Attribute{
				Description: "RADIUS accounting port. Defaults to 1813.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1813),
			},
			"x_secret": schema.StringAttribute{
				Description: "RADIUS shared secret (write-only, 1-48 characters).",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tunneled_reply": schema.BoolAttribute{
				Description: "Enable tunneled reply.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"interim_update_interval": schema.Int64Attribute{
				Description: "Interim update interval in seconds (60-86400).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 86400),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *SettingRadiusResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *SettingRadiusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingRadiusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	setting := r.planToSDK(&plan)
	savedSecret := plan.XSecret

	updated, err := r.client.UpdateSettingRadius(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "RADIUS setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XSecret = savedSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingRadiusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingRadiusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, 2*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	savedSecret := state.XSecret

	setting, err := r.client.GetSettingRadius(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "RADIUS setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(setting, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.XSecret.IsNull() || state.XSecret.ValueString() == "" {
		state.XSecret = savedSecret
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingRadiusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingRadiusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	setting := r.planToSDK(&plan)
	if !plan.ID.IsNull() {
		setting.ID = plan.ID.ValueString()
	}
	savedSecret := plan.XSecret

	updated, err := r.client.UpdateSettingRadius(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "RADIUS setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XSecret = savedSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingRadiusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingRadiusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	defaults := &unifi.SettingRadius{
		Enabled:           boolPtr(false),
		AccountingEnabled: boolPtr(false),
		AuthPort:          intPtr(1812),
		AcctPort:          intPtr(1813),
		TunneledReply:     boolPtr(false),
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingRadius(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "RADIUS setting")
	}
}

func (r *SettingRadiusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingRadiusResource) planToSDK(plan *SettingRadiusResourceModel) *unifi.SettingRadius {
	s := &unifi.SettingRadius{
		Key:               "radius",
		Enabled:           boolPtr(plan.Enabled.ValueBool()),
		AccountingEnabled: boolPtr(plan.AccountingEnabled.ValueBool()),
		TunneledReply:     boolPtr(plan.TunneledReply.ValueBool()),
	}

	if !plan.AuthPort.IsNull() && !plan.AuthPort.IsUnknown() {
		s.AuthPort = intPtr(plan.AuthPort.ValueInt64())
	}
	if !plan.AcctPort.IsNull() && !plan.AcctPort.IsUnknown() {
		s.AcctPort = intPtr(plan.AcctPort.ValueInt64())
	}
	if !plan.XSecret.IsNull() && !plan.XSecret.IsUnknown() {
		s.XSecret = plan.XSecret.ValueString()
	}
	if !plan.InterimUpdateInterval.IsNull() && !plan.InterimUpdateInterval.IsUnknown() {
		s.InterimUpdateInterval = intPtr(plan.InterimUpdateInterval.ValueInt64())
	}

	return s
}

func (r *SettingRadiusResource) sdkToState(setting *unifi.SettingRadius, state *SettingRadiusResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.Enabled = types.BoolValue(derefBool(setting.Enabled))
	state.AccountingEnabled = types.BoolValue(derefBool(setting.AccountingEnabled))
	state.TunneledReply = types.BoolValue(derefBool(setting.TunneledReply))

	if setting.AuthPort != nil {
		state.AuthPort = types.Int64Value(int64(*setting.AuthPort))
	}
	if setting.AcctPort != nil {
		state.AcctPort = types.Int64Value(int64(*setting.AcctPort))
	}
	if setting.InterimUpdateInterval != nil {
		state.InterimUpdateInterval = types.Int64Value(int64(*setting.InterimUpdateInterval))
	} else {
		state.InterimUpdateInterval = types.Int64Null()
	}

	// x_secret is write-only — never read from API.
	// Callers (Create/Update) restore it from plan; Read preserves prior state.

	return diags
}
