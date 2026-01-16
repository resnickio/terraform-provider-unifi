package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &PortProfileResource{}
	_ resource.ResourceWithImportState = &PortProfileResource{}
)

type PortProfileResource struct {
	client *AutoLoginClient
}

type PortProfileResourceModel struct {
	ID     types.String `tfsdk:"id"`
	SiteID types.String `tfsdk:"site_id"`
	Name   types.String `tfsdk:"name"`

	NativeNetworkID    types.String `tfsdk:"native_network_id"`
	TaggedVlanMgmt     types.String `tfsdk:"tagged_vlan_mgmt"`
	ExcludedNetworkIDs types.Set    `tfsdk:"excluded_network_ids"`

	Autoneg    types.Bool  `tfsdk:"autoneg"`
	Speed      types.Int64 `tfsdk:"speed"`
	FullDuplex types.Bool  `tfsdk:"full_duplex"`

	PoeMode types.String `tfsdk:"poe_mode"`

	OpMode    types.String `tfsdk:"op_mode"`
	Isolation types.Bool   `tfsdk:"isolation"`

	Dot1xCtrl        types.String `tfsdk:"dot1x_ctrl"`
	Dot1xIdleTimeout types.Int64  `tfsdk:"dot1x_idle_timeout"`

	StpPortMode          types.Bool `tfsdk:"stp_port_mode"`
	LldpmedEnabled       types.Bool `tfsdk:"lldpmed_enabled"`
	LldpmedNotifyEnabled types.Bool `tfsdk:"lldpmed_notify_enabled"`

	StormctrlBcastEnabled types.Bool  `tfsdk:"stormctrl_bcast_enabled"`
	StormctrlBcastRate    types.Int64 `tfsdk:"stormctrl_bcast_rate"`
	StormctrlMcastEnabled types.Bool  `tfsdk:"stormctrl_mcast_enabled"`
	StormctrlMcastRate    types.Int64 `tfsdk:"stormctrl_mcast_rate"`
	StormctrlUcastEnabled types.Bool  `tfsdk:"stormctrl_ucast_enabled"`
	StormctrlUcastRate    types.Int64 `tfsdk:"stormctrl_ucast_rate"`

	EgressRateLimitKbpsEnabled types.Bool  `tfsdk:"egress_rate_limit_kbps_enabled"`
	EgressRateLimitKbps        types.Int64 `tfsdk:"egress_rate_limit_kbps"`

	PortSecurityEnabled    types.Bool `tfsdk:"port_security_enabled"`
	PortSecurityMacAddress types.Set  `tfsdk:"port_security_mac_address"`

	PortKeepaliveEnabled types.Bool `tfsdk:"port_keepalive_enabled"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func NewPortProfileResource() resource.Resource {
	return &PortProfileResource{}
}

func (r *PortProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_profile"
}

func (r *PortProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi switch port profile (PortConf) for configuring switch port settings including VLANs, PoE, 802.1X, and storm control.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the port profile.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the port profile is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the port profile.",
				Required:    true,
			},

			"native_network_id": schema.StringAttribute{
				Description: "Network ID for the native/untagged VLAN.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tagged_vlan_mgmt": schema.StringAttribute{
				Description: "Tagged VLAN management mode. Valid values: 'all' (allow all VLANs), 'block' (block all VLANs), 'custom' (allow all except excluded).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "block", "custom"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"excluded_network_ids": schema.SetAttribute{
				Description: "Set of network IDs to exclude from tagged VLANs. Only applicable when tagged_vlan_mgmt is 'custom'. Networks not in this list will be allowed as tagged VLANs.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"autoneg": schema.BoolAttribute{
				Description: "Enable auto-negotiation for link speed and duplex.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"speed": schema.Int64Attribute{
				Description: "Link speed in Mbps (10, 100, 1000, 2500, 5000, 10000). Only applicable when autoneg is disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"full_duplex": schema.BoolAttribute{
				Description: "Enable full duplex mode. Only applicable when autoneg is disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			"poe_mode": schema.StringAttribute{
				Description: "PoE mode. Valid values: 'auto', 'pasv24', 'passthrough', 'off'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "pasv24", "passthrough", "off"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"op_mode": schema.StringAttribute{
				Description: "Port operation mode. Valid values: 'switch', 'mirror', 'aggregate'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("switch", "mirror", "aggregate"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"isolation": schema.BoolAttribute{
				Description: "Enable port isolation to prevent traffic between ports using this profile.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			"dot1x_ctrl": schema.StringAttribute{
				Description: "802.1X control mode. Valid values: 'force_authorized', 'force_unauthorized', 'auto', 'mac_based', 'multi_host'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("force_authorized", "force_unauthorized", "auto", "mac_based", "multi_host"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dot1x_idle_timeout": schema.Int64Attribute{
				Description: "802.1X idle timeout in seconds.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"stp_port_mode": schema.BoolAttribute{
				Description: "Enable STP (Spanning Tree Protocol) on this port.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"lldpmed_enabled": schema.BoolAttribute{
				Description: "Enable LLDP-MED (Link Layer Discovery Protocol - Media Endpoint Discovery).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"lldpmed_notify_enabled": schema.BoolAttribute{
				Description: "Enable LLDP-MED topology change notifications.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			"stormctrl_bcast_enabled": schema.BoolAttribute{
				Description: "Enable broadcast storm control.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"stormctrl_bcast_rate": schema.Int64Attribute{
				Description: "Broadcast storm control rate limit.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"stormctrl_mcast_enabled": schema.BoolAttribute{
				Description: "Enable multicast storm control.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"stormctrl_mcast_rate": schema.Int64Attribute{
				Description: "Multicast storm control rate limit.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"stormctrl_ucast_enabled": schema.BoolAttribute{
				Description: "Enable unicast storm control.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"stormctrl_ucast_rate": schema.Int64Attribute{
				Description: "Unicast storm control rate limit.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"egress_rate_limit_kbps_enabled": schema.BoolAttribute{
				Description: "Enable egress rate limiting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"egress_rate_limit_kbps": schema.Int64Attribute{
				Description: "Egress rate limit in Kbps.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"port_security_enabled": schema.BoolAttribute{
				Description: "Enable port security (MAC address filtering).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"port_security_mac_address": schema.SetAttribute{
				Description: "Set of allowed MAC addresses when port security is enabled.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"port_keepalive_enabled": schema.BoolAttribute{
				Description: "Enable port keepalive.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
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

func (r *PortProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *PortProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PortProfileResourceModel

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

	profile := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePortProfile(ctx, profile)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "port profile")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PortProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PortProfileResourceModel

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

	profile, err := r.client.GetPortProfile(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "port profile")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, profile, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PortProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PortProfileResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PortProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	profile := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	profile.ID = state.ID.ValueString()
	profile.SiteID = state.SiteID.ValueString()

	_, err := r.client.UpdatePortProfile(ctx, state.ID.ValueString(), profile)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "port profile")
		return
	}

	updated, err := r.client.GetPortProfile(ctx, state.ID.ValueString())
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "port profile")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PortProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PortProfileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	err := r.client.DeletePortProfile(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "port profile")
		return
	}
}

func (r *PortProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PortProfileResource) planToSDK(ctx context.Context, plan *PortProfileResourceModel, diags *diag.Diagnostics) *unifi.PortConf {
	profile := &unifi.PortConf{
		Name: plan.Name.ValueString(),
	}

	if !plan.NativeNetworkID.IsNull() && !plan.NativeNetworkID.IsUnknown() {
		profile.NativeNetworkconfID = plan.NativeNetworkID.ValueString()
	}
	if !plan.TaggedVlanMgmt.IsNull() && !plan.TaggedVlanMgmt.IsUnknown() {
		profile.TaggedVlanMgmt = plan.TaggedVlanMgmt.ValueString()
	}
	if !plan.ExcludedNetworkIDs.IsNull() && !plan.ExcludedNetworkIDs.IsUnknown() {
		var excludedIDs []string
		diags.Append(plan.ExcludedNetworkIDs.ElementsAs(ctx, &excludedIDs, false)...)
		if diags.HasError() {
			return nil
		}
		profile.ExcludedNetworkconfIDs = excludedIDs
	}

	if !plan.Autoneg.IsNull() && !plan.Autoneg.IsUnknown() {
		profile.Autoneg = boolPtr(plan.Autoneg.ValueBool())
	}
	if !plan.Speed.IsNull() && !plan.Speed.IsUnknown() {
		profile.Speed = intPtr(plan.Speed.ValueInt64())
	}
	if !plan.FullDuplex.IsNull() && !plan.FullDuplex.IsUnknown() {
		profile.FullDuplex = boolPtr(plan.FullDuplex.ValueBool())
	}

	if !plan.PoeMode.IsNull() && !plan.PoeMode.IsUnknown() {
		profile.PoeMode = plan.PoeMode.ValueString()
	}

	if !plan.OpMode.IsNull() && !plan.OpMode.IsUnknown() {
		profile.OpMode = plan.OpMode.ValueString()
	}
	if !plan.Isolation.IsNull() && !plan.Isolation.IsUnknown() {
		profile.Isolation = boolPtr(plan.Isolation.ValueBool())
	}

	if !plan.Dot1xCtrl.IsNull() && !plan.Dot1xCtrl.IsUnknown() {
		profile.Dot1xCtrl = plan.Dot1xCtrl.ValueString()
	}
	if !plan.Dot1xIdleTimeout.IsNull() && !plan.Dot1xIdleTimeout.IsUnknown() {
		profile.Dot1xIDleTimeout = intPtr(plan.Dot1xIdleTimeout.ValueInt64())
	}

	if !plan.StpPortMode.IsNull() && !plan.StpPortMode.IsUnknown() {
		profile.StpPortMode = boolPtr(plan.StpPortMode.ValueBool())
	}
	if !plan.LldpmedEnabled.IsNull() && !plan.LldpmedEnabled.IsUnknown() {
		profile.LldpmedEnabled = boolPtr(plan.LldpmedEnabled.ValueBool())
	}
	if !plan.LldpmedNotifyEnabled.IsNull() && !plan.LldpmedNotifyEnabled.IsUnknown() {
		profile.LldpmedNotifyEnabled = boolPtr(plan.LldpmedNotifyEnabled.ValueBool())
	}

	if !plan.StormctrlBcastEnabled.IsNull() && !plan.StormctrlBcastEnabled.IsUnknown() {
		profile.StormctrlBcastEnabled = boolPtr(plan.StormctrlBcastEnabled.ValueBool())
	}
	if !plan.StormctrlBcastRate.IsNull() && !plan.StormctrlBcastRate.IsUnknown() {
		profile.StormctrlBcastRate = intPtr(plan.StormctrlBcastRate.ValueInt64())
	}
	if !plan.StormctrlMcastEnabled.IsNull() && !plan.StormctrlMcastEnabled.IsUnknown() {
		profile.StormctrlMcastEnabled = boolPtr(plan.StormctrlMcastEnabled.ValueBool())
	}
	if !plan.StormctrlMcastRate.IsNull() && !plan.StormctrlMcastRate.IsUnknown() {
		profile.StormctrlMcastRate = intPtr(plan.StormctrlMcastRate.ValueInt64())
	}
	if !plan.StormctrlUcastEnabled.IsNull() && !plan.StormctrlUcastEnabled.IsUnknown() {
		profile.StormctrlUcastEnabled = boolPtr(plan.StormctrlUcastEnabled.ValueBool())
	}
	if !plan.StormctrlUcastRate.IsNull() && !plan.StormctrlUcastRate.IsUnknown() {
		profile.StormctrlUcastRate = intPtr(plan.StormctrlUcastRate.ValueInt64())
	}

	if !plan.EgressRateLimitKbpsEnabled.IsNull() && !plan.EgressRateLimitKbpsEnabled.IsUnknown() {
		profile.EgressRateLimitEnabled = boolPtr(plan.EgressRateLimitKbpsEnabled.ValueBool())
	}
	if !plan.EgressRateLimitKbps.IsNull() && !plan.EgressRateLimitKbps.IsUnknown() {
		profile.EgressRateLimitKbps = intPtr(plan.EgressRateLimitKbps.ValueInt64())
	}

	if !plan.PortSecurityEnabled.IsNull() && !plan.PortSecurityEnabled.IsUnknown() {
		profile.PortSecurityEnabled = boolPtr(plan.PortSecurityEnabled.ValueBool())
	}
	if !plan.PortSecurityMacAddress.IsNull() && !plan.PortSecurityMacAddress.IsUnknown() {
		var macs []string
		diags.Append(plan.PortSecurityMacAddress.ElementsAs(ctx, &macs, false)...)
		if diags.HasError() {
			return nil
		}
		profile.PortSecurityMacAddress = macs
	}

	if !plan.PortKeepaliveEnabled.IsNull() && !plan.PortKeepaliveEnabled.IsUnknown() {
		profile.PortKeepaliveEnabled = boolPtr(plan.PortKeepaliveEnabled.ValueBool())
	}

	return profile
}

func (r *PortProfileResource) sdkToState(ctx context.Context, profile *unifi.PortConf, state *PortProfileResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(profile.ID)
	state.SiteID = types.StringValue(profile.SiteID)
	state.Name = types.StringValue(profile.Name)

	state.NativeNetworkID = stringValueOrNull(profile.NativeNetworkconfID)
	state.TaggedVlanMgmt = stringValueOrNull(profile.TaggedVlanMgmt)

	if profile.ExcludedNetworkconfIDs == nil {
		profile.ExcludedNetworkconfIDs = []string{}
	}
	excludedIDs, d := types.SetValueFrom(ctx, types.StringType, profile.ExcludedNetworkconfIDs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	state.ExcludedNetworkIDs = excludedIDs

	if profile.Autoneg != nil {
		state.Autoneg = types.BoolValue(*profile.Autoneg)
	} else {
		state.Autoneg = types.BoolNull()
	}
	if profile.Speed != nil {
		state.Speed = types.Int64Value(int64(*profile.Speed))
	} else {
		state.Speed = types.Int64Null()
	}
	if profile.FullDuplex != nil {
		state.FullDuplex = types.BoolValue(*profile.FullDuplex)
	} else {
		state.FullDuplex = types.BoolNull()
	}

	state.PoeMode = stringValueOrNull(profile.PoeMode)
	state.OpMode = stringValueOrNull(profile.OpMode)

	if profile.Isolation != nil {
		state.Isolation = types.BoolValue(*profile.Isolation)
	} else {
		state.Isolation = types.BoolNull()
	}

	state.Dot1xCtrl = stringValueOrNull(profile.Dot1xCtrl)
	if profile.Dot1xIDleTimeout != nil {
		state.Dot1xIdleTimeout = types.Int64Value(int64(*profile.Dot1xIDleTimeout))
	} else {
		state.Dot1xIdleTimeout = types.Int64Null()
	}

	if profile.StpPortMode != nil {
		state.StpPortMode = types.BoolValue(*profile.StpPortMode)
	} else {
		state.StpPortMode = types.BoolNull()
	}
	if profile.LldpmedEnabled != nil {
		state.LldpmedEnabled = types.BoolValue(*profile.LldpmedEnabled)
	} else {
		state.LldpmedEnabled = types.BoolNull()
	}
	if profile.LldpmedNotifyEnabled != nil {
		state.LldpmedNotifyEnabled = types.BoolValue(*profile.LldpmedNotifyEnabled)
	} else {
		state.LldpmedNotifyEnabled = types.BoolNull()
	}

	if profile.StormctrlBcastEnabled != nil {
		state.StormctrlBcastEnabled = types.BoolValue(*profile.StormctrlBcastEnabled)
	} else {
		state.StormctrlBcastEnabled = types.BoolNull()
	}
	if profile.StormctrlBcastRate != nil {
		state.StormctrlBcastRate = types.Int64Value(int64(*profile.StormctrlBcastRate))
	} else {
		state.StormctrlBcastRate = types.Int64Null()
	}
	if profile.StormctrlMcastEnabled != nil {
		state.StormctrlMcastEnabled = types.BoolValue(*profile.StormctrlMcastEnabled)
	} else {
		state.StormctrlMcastEnabled = types.BoolNull()
	}
	if profile.StormctrlMcastRate != nil {
		state.StormctrlMcastRate = types.Int64Value(int64(*profile.StormctrlMcastRate))
	} else {
		state.StormctrlMcastRate = types.Int64Null()
	}
	if profile.StormctrlUcastEnabled != nil {
		state.StormctrlUcastEnabled = types.BoolValue(*profile.StormctrlUcastEnabled)
	} else {
		state.StormctrlUcastEnabled = types.BoolNull()
	}
	if profile.StormctrlUcastRate != nil {
		state.StormctrlUcastRate = types.Int64Value(int64(*profile.StormctrlUcastRate))
	} else {
		state.StormctrlUcastRate = types.Int64Null()
	}

	if profile.EgressRateLimitEnabled != nil {
		state.EgressRateLimitKbpsEnabled = types.BoolValue(*profile.EgressRateLimitEnabled)
	} else {
		state.EgressRateLimitKbpsEnabled = types.BoolNull()
	}
	if profile.EgressRateLimitKbps != nil {
		state.EgressRateLimitKbps = types.Int64Value(int64(*profile.EgressRateLimitKbps))
	} else {
		state.EgressRateLimitKbps = types.Int64Null()
	}

	if profile.PortSecurityEnabled != nil {
		state.PortSecurityEnabled = types.BoolValue(*profile.PortSecurityEnabled)
	} else {
		state.PortSecurityEnabled = types.BoolNull()
	}
	if profile.PortSecurityMacAddress == nil {
		profile.PortSecurityMacAddress = []string{}
	}
	macs, d := types.SetValueFrom(ctx, types.StringType, profile.PortSecurityMacAddress)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	state.PortSecurityMacAddress = macs

	if profile.PortKeepaliveEnabled != nil {
		state.PortKeepaliveEnabled = types.BoolValue(*profile.PortKeepaliveEnabled)
	} else {
		state.PortKeepaliveEnabled = types.BoolNull()
	}

	return diags
}
