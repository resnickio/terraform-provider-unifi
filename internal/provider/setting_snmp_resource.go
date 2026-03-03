package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &SettingSNMPResource{}
	_ resource.ResourceWithImportState = &SettingSNMPResource{}
)

type SettingSNMPResource struct {
	client *AutoLoginClient
}

type SettingSNMPResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	SiteID    types.String   `tfsdk:"site_id"`
	Enabled   types.Bool     `tfsdk:"enabled"`
	Community types.String   `tfsdk:"community"`
	EnabledV3 types.Bool     `tfsdk:"enabled_v3"`
	Username  types.String   `tfsdk:"username"`
	XPassword types.String   `tfsdk:"x_password"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func NewSettingSNMPResource() resource.Resource {
	return &SettingSNMPResource{}
}

func (r *SettingSNMPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_snmp"
}

func (r *SettingSNMPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi SNMP settings. " +
			"This is a singleton resource — one per site. Delete resets to defaults.",
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
				Description: "Enable SNMP.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"community": schema.StringAttribute{
				Description: "SNMP community string (v1/v2c).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled_v3": schema.BoolAttribute{
				Description: "Enable SNMPv3.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"username": schema.StringAttribute{
				Description: "SNMPv3 username.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"x_password": schema.StringAttribute{
				Description: "SNMPv3 password (write-only).",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *SettingSNMPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingSNMPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingSNMPResourceModel
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
	savedPassword := plan.XPassword

	updated, err := r.client.UpdateSettingSNMP(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "SNMP setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XPassword = savedPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingSNMPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingSNMPResourceModel
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

	savedPassword := state.XPassword

	setting, err := r.client.GetSettingSNMP(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "SNMP setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(setting, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.XPassword.IsNull() || state.XPassword.ValueString() == "" {
		state.XPassword = savedPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingSNMPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingSNMPResourceModel
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

	savedPassword := plan.XPassword

	updated, err := r.client.UpdateSettingSNMP(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "SNMP setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XPassword = savedPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingSNMPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingSNMPResourceModel
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

	defaults := &unifi.SettingSNMP{
		Key:       "snmp",
		Enabled:   boolPtr(false),
		EnabledV3: boolPtr(false),
		Community: "",
		Username:  "",
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingSNMP(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "SNMP setting")
		return
	}
}

func (r *SettingSNMPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingSNMPResource) planToSDK(plan *SettingSNMPResourceModel) *unifi.SettingSNMP {
	s := &unifi.SettingSNMP{
		Key:       "snmp",
		Enabled:   boolPtr(plan.Enabled.ValueBool()),
		EnabledV3: boolPtr(plan.EnabledV3.ValueBool()),
	}

	if !plan.Community.IsNull() && !plan.Community.IsUnknown() {
		s.Community = plan.Community.ValueString()
	}
	if !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		s.Username = plan.Username.ValueString()
	}
	if !plan.XPassword.IsNull() && !plan.XPassword.IsUnknown() {
		s.XPassword = plan.XPassword.ValueString()
	}

	return s
}

func (r *SettingSNMPResource) sdkToState(setting *unifi.SettingSNMP, state *SettingSNMPResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.Enabled = types.BoolValue(derefBool(setting.Enabled))
	state.Community = stringValueOrNull(setting.Community)
	state.EnabledV3 = types.BoolValue(derefBool(setting.EnabledV3))
	state.Username = stringValueOrNull(setting.Username)

	return diags
}
