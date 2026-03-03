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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &SettingMgmtResource{}
	_ resource.ResourceWithImportState = &SettingMgmtResource{}
)

type SettingMgmtResource struct {
	client *AutoLoginClient
}

type SettingMgmtResourceModel struct {
	ID                      types.String   `tfsdk:"id"`
	SiteID                  types.String   `tfsdk:"site_id"`
	AutoUpgrade             types.Bool     `tfsdk:"auto_upgrade"`
	AutoUpgradeHour         types.Int64    `tfsdk:"auto_upgrade_hour"`
	LEDEnabled              types.Bool     `tfsdk:"led_enabled"`
	AlertEnabled            types.Bool     `tfsdk:"alert_enabled"`
	XSSHEnabled             types.Bool     `tfsdk:"x_ssh_enabled"`
	XSSHAuthPasswordEnabled types.Bool     `tfsdk:"x_ssh_auth_password_enabled"`
	XSSHUsername            types.String   `tfsdk:"x_ssh_username"`
	XSSHPassword            types.String   `tfsdk:"x_ssh_password"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

func NewSettingMgmtResource() resource.Resource {
	return &SettingMgmtResource{}
}

func (r *SettingMgmtResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_mgmt"
}

func (r *SettingMgmtResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi site management settings (auto-upgrade, LED, SSH, alerts). " +
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
			"auto_upgrade": schema.BoolAttribute{
				Description: "Enable automatic device firmware upgrades.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"auto_upgrade_hour": schema.Int64Attribute{
				Description: "Hour of day for auto-upgrades (0-23).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"led_enabled": schema.BoolAttribute{
				Description: "Enable device LEDs.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"alert_enabled": schema.BoolAttribute{
				Description: "Enable alerts.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"x_ssh_enabled": schema.BoolAttribute{
				Description: "Enable SSH access to devices.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"x_ssh_auth_password_enabled": schema.BoolAttribute{
				Description: "Enable SSH password authentication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"x_ssh_username": schema.StringAttribute{
				Description: "SSH username.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"x_ssh_password": schema.StringAttribute{
				Description: "SSH password (write-only).",
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

func (r *SettingMgmtResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingMgmtResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingMgmtResourceModel
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
	savedPassword := plan.XSSHPassword

	updated, err := r.client.UpdateSettingMgmt(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "management setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XSSHPassword = savedPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingMgmtResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingMgmtResourceModel
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

	savedPassword := state.XSSHPassword

	setting, err := r.client.GetSettingMgmt(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "management setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(setting, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.XSSHPassword.IsNull() || state.XSSHPassword.ValueString() == "" {
		state.XSSHPassword = savedPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingMgmtResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingMgmtResourceModel
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

	savedPassword := plan.XSSHPassword

	updated, err := r.client.UpdateSettingMgmt(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "management setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XSSHPassword = savedPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingMgmtResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingMgmtResourceModel
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

	defaults := &unifi.SettingMgmt{
		AutoUpgrade:             boolPtr(false),
		LEDEnabled:              boolPtr(true),
		AlertEnabled:            boolPtr(true),
		XSSHEnabled:             boolPtr(false),
		XSSHAuthPasswordEnabled: boolPtr(false),
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingMgmt(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "management setting")
		return
	}
}

func (r *SettingMgmtResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingMgmtResource) planToSDK(plan *SettingMgmtResourceModel) *unifi.SettingMgmt {
	s := &unifi.SettingMgmt{
		Key:                     "mgmt",
		AutoUpgrade:             boolPtr(plan.AutoUpgrade.ValueBool()),
		LEDEnabled:              boolPtr(plan.LEDEnabled.ValueBool()),
		AlertEnabled:            boolPtr(plan.AlertEnabled.ValueBool()),
		XSSHEnabled:             boolPtr(plan.XSSHEnabled.ValueBool()),
		XSSHAuthPasswordEnabled: boolPtr(plan.XSSHAuthPasswordEnabled.ValueBool()),
	}

	if !plan.AutoUpgradeHour.IsNull() && !plan.AutoUpgradeHour.IsUnknown() {
		v := int(plan.AutoUpgradeHour.ValueInt64())
		s.AutoUpgradeHour = &v
	}
	if !plan.XSSHUsername.IsNull() && !plan.XSSHUsername.IsUnknown() {
		s.XSSHUsername = plan.XSSHUsername.ValueString()
	}
	if !plan.XSSHPassword.IsNull() && !plan.XSSHPassword.IsUnknown() {
		s.XSSHPassword = plan.XSSHPassword.ValueString()
	}

	return s
}

func (r *SettingMgmtResource) sdkToState(setting *unifi.SettingMgmt, state *SettingMgmtResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.AutoUpgrade = types.BoolValue(derefBool(setting.AutoUpgrade))
	state.LEDEnabled = types.BoolValue(derefBool(setting.LEDEnabled))
	state.AlertEnabled = types.BoolValue(derefBool(setting.AlertEnabled))
	state.XSSHEnabled = types.BoolValue(derefBool(setting.XSSHEnabled))
	state.XSSHAuthPasswordEnabled = types.BoolValue(derefBool(setting.XSSHAuthPasswordEnabled))
	state.XSSHUsername = stringValueOrNull(setting.XSSHUsername)

	if setting.AutoUpgradeHour != nil {
		state.AutoUpgradeHour = types.Int64Value(int64(*setting.AutoUpgradeHour))
	} else {
		state.AutoUpgradeHour = types.Int64Null()
	}

	// x_ssh_password is write-only — never read from API.
	// Callers (Create/Update) restore it from plan; Read preserves prior state.

	return diags
}
