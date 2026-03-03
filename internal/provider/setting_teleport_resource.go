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
	_ resource.Resource                = &SettingTeleportResource{}
	_ resource.ResourceWithImportState = &SettingTeleportResource{}
)

type SettingTeleportResource struct {
	client *AutoLoginClient
}

type SettingTeleportResourceModel struct {
	ID         types.String   `tfsdk:"id"`
	SiteID     types.String   `tfsdk:"site_id"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	SubnetCIDR types.String   `tfsdk:"subnet_cidr"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

func NewSettingTeleportResource() resource.Resource {
	return &SettingTeleportResource{}
}

func (r *SettingTeleportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_teleport"
}

func (r *SettingTeleportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi Teleport settings. " +
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
				Description: "Enable Teleport.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"subnet_cidr": schema.StringAttribute{
				Description: "Subnet CIDR for Teleport VPN clients.",
				Optional:    true,
				Computed:    true,
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

func (r *SettingTeleportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingTeleportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingTeleportResourceModel
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

	updated, err := r.client.UpdateSettingTeleport(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "teleport setting")
		return
	}

	savedSubnetCIDR := plan.SubnetCIDR
	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.SubnetCIDR.IsNull() && !savedSubnetCIDR.IsNull() {
		plan.SubnetCIDR = savedSubnetCIDR
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingTeleportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingTeleportResourceModel
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

	savedSubnetCIDR := state.SubnetCIDR

	setting, err := r.client.GetSettingTeleport(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "teleport setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(setting, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.SubnetCIDR.IsNull() && !savedSubnetCIDR.IsNull() {
		state.SubnetCIDR = savedSubnetCIDR
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingTeleportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingTeleportResourceModel
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

	updated, err := r.client.UpdateSettingTeleport(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "teleport setting")
		return
	}

	savedSubnetCIDR := plan.SubnetCIDR
	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.SubnetCIDR.IsNull() && !savedSubnetCIDR.IsNull() {
		plan.SubnetCIDR = savedSubnetCIDR
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingTeleportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingTeleportResourceModel
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

	defaults := &unifi.SettingTeleport{
		Key:     "teleport",
		Enabled: boolPtr(false),
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingTeleport(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "teleport setting")
		return
	}
}

func (r *SettingTeleportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingTeleportResource) planToSDK(plan *SettingTeleportResourceModel) *unifi.SettingTeleport {
	s := &unifi.SettingTeleport{
		Key:     "teleport",
		Enabled: boolPtr(plan.Enabled.ValueBool()),
	}

	if !plan.SubnetCIDR.IsNull() && !plan.SubnetCIDR.IsUnknown() {
		s.SubnetCIDR = plan.SubnetCIDR.ValueString()
	}

	return s
}

func (r *SettingTeleportResource) sdkToState(setting *unifi.SettingTeleport, state *SettingTeleportResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.Enabled = types.BoolValue(derefBool(setting.Enabled))
	state.SubnetCIDR = stringValueOrNull(setting.SubnetCIDR)

	return diags
}
