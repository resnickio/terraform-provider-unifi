package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &DevicePortOverrideResource{}
	_ resource.ResourceWithImportState = &DevicePortOverrideResource{}
)

type DevicePortOverrideResource struct {
	client *AutoLoginClient
}

type DevicePortOverrideResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	DeviceID              types.String `tfsdk:"device_id"`
	MAC                   types.String `tfsdk:"mac"`
	PortIdx               types.Int64  `tfsdk:"port_idx"`
	Name                  types.String `tfsdk:"name"`
	PortProfileID         types.String `tfsdk:"port_profile_id"`
	PoeMode               types.String `tfsdk:"poe_mode"`
	OpMode                types.String `tfsdk:"op_mode"`
	AggregateMembers      types.Set    `tfsdk:"aggregate_members"`
	NativeNetworkID       types.String `tfsdk:"native_network_id"`
	TaggedNetworkIDs      types.Set    `tfsdk:"tagged_network_ids"`
	ExcludedNetworkIDs    types.Set    `tfsdk:"excluded_network_ids"`
	VoiceNetworkID        types.String `tfsdk:"voice_network_id"`
	Autoneg               types.Bool   `tfsdk:"autoneg"`
	Speed                 types.Int64  `tfsdk:"speed"`
	FullDuplex            types.Bool   `tfsdk:"full_duplex"`
	Isolation             types.Bool   `tfsdk:"isolation"`
	StpPortMode           types.Bool   `tfsdk:"stp_port_mode"`
	EgressRateLimitKbps   types.Int64  `tfsdk:"egress_rate_limit_kbps"`
	PortSecurityEnabled   types.Bool   `tfsdk:"port_security_enabled"`
	PortSecurityMacAddresses types.Set `tfsdk:"port_security_mac_addresses"`
}

func NewDevicePortOverrideResource() resource.Resource {
	return &DevicePortOverrideResource{}
}

func (r *DevicePortOverrideResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_port_override"
}

func (r *DevicePortOverrideResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages port-specific configuration overrides on a UniFi device (switch). Use this to assign port profiles, configure PoE, set port names, and other per-port settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier (device_id:port_idx).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Description: "The ID of the device. Use the unifi_device data source to look this up.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac": schema.StringAttribute{
				Description: "The MAC address of the device (computed from device_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"port_idx": schema.Int64Attribute{
				Description: "The port index (1-based). Port 1 is port_idx=1.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name/label for this port.",
				Optional:    true,
			},
			"port_profile_id": schema.StringAttribute{
				Description: "The port profile ID to apply to this port. Use unifi_port_profile resource or data source.",
				Optional:    true,
			},
			"poe_mode": schema.StringAttribute{
				Description: "PoE mode for this port. Valid values: 'auto', 'off', 'pasv24', 'passthrough'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "off", "pasv24", "passthrough"),
				},
			},
			"op_mode": schema.StringAttribute{
				Description: "Operation mode for this port. Valid values: 'switch', 'mirror', 'aggregate'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("switch", "mirror", "aggregate"),
				},
			},
			"aggregate_members": schema.SetAttribute{
				Description: "Port indices to include in link aggregation (when op_mode is 'aggregate').",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"native_network_id": schema.StringAttribute{
				Description: "The native (untagged) network ID for this port.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tagged_network_ids": schema.SetAttribute{
				Description: "Set of tagged network IDs for this port.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"excluded_network_ids": schema.SetAttribute{
				Description: "Set of excluded network IDs for this port.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"voice_network_id": schema.StringAttribute{
				Description: "The voice network ID for this port.",
				Optional:    true,
			},
			"autoneg": schema.BoolAttribute{
				Description: "Enable auto-negotiation for speed and duplex.",
				Optional:    true,
			},
			"speed": schema.Int64Attribute{
				Description: "Port speed in Mbps (when autoneg is disabled). Valid values: 10, 100, 1000, 2500, 10000.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.OneOf(10, 100, 1000, 2500, 10000),
				},
			},
			"full_duplex": schema.BoolAttribute{
				Description: "Enable full duplex (when autoneg is disabled).",
				Optional:    true,
			},
			"isolation": schema.BoolAttribute{
				Description: "Enable port isolation.",
				Optional:    true,
			},
			"stp_port_mode": schema.BoolAttribute{
				Description: "Enable Spanning Tree Protocol on this port.",
				Optional:    true,
			},
			"egress_rate_limit_kbps": schema.Int64Attribute{
				Description: "Egress rate limit in Kbps. Set to 0 to disable.",
				Optional:    true,
			},
			"port_security_enabled": schema.BoolAttribute{
				Description: "Enable port security (MAC address limiting).",
				Optional:    true,
			},
			"port_security_mac_addresses": schema.SetAttribute{
				Description: "Set of allowed MAC addresses when port security is enabled.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *DevicePortOverrideResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DevicePortOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DevicePortOverrideResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device, err := r.getDeviceByID(ctx, plan.DeviceID.ValueString())
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "device")
		return
	}

	portIdx := int(plan.PortIdx.ValueInt64())
	override := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	device.PortOverrides = r.mergePortOverride(device.PortOverrides, override, portIdx)

	updated, err := r.client.UpdateDevice(ctx, device.ID, device)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "device port override")
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%d", device.ID, portIdx))
	plan.MAC = types.StringValue(updated.MAC)

	foundOverride := r.findPortOverride(updated.PortOverrides, portIdx)
	if foundOverride != nil {
		resp.Diagnostics.Append(r.sdkToState(ctx, foundOverride, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DevicePortOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DevicePortOverrideResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device, err := r.getDeviceByID(ctx, state.DeviceID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "device")
		return
	}

	portIdx := int(state.PortIdx.ValueInt64())
	override := r.findPortOverride(device.PortOverrides, portIdx)

	if override == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.MAC = types.StringValue(device.MAC)
	resp.Diagnostics.Append(r.sdkToState(ctx, override, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DevicePortOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DevicePortOverrideResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device, err := r.getDeviceByID(ctx, plan.DeviceID.ValueString())
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "device")
		return
	}

	portIdx := int(plan.PortIdx.ValueInt64())
	override := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	device.PortOverrides = r.mergePortOverride(device.PortOverrides, override, portIdx)

	updated, err := r.client.UpdateDevice(ctx, device.ID, device)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "device port override")
		return
	}

	plan.MAC = types.StringValue(updated.MAC)

	foundOverride := r.findPortOverride(updated.PortOverrides, portIdx)
	if foundOverride != nil {
		resp.Diagnostics.Append(r.sdkToState(ctx, foundOverride, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DevicePortOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DevicePortOverrideResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	device, err := r.getDeviceByID(ctx, state.DeviceID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "device")
		return
	}

	portIdx := int(state.PortIdx.ValueInt64())
	device.PortOverrides = r.removePortOverride(device.PortOverrides, portIdx)

	_, err = r.client.UpdateDevice(ctx, device.ID, device)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "delete", "device port override")
		return
	}
}

func (r *DevicePortOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format 'device_id:port_idx', got '%s'", req.ID),
		)
		return
	}

	deviceID := parts[0]
	portIdx, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("port_idx must be a number, got '%s'", parts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), deviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port_idx"), portIdx)...)
}

func (r *DevicePortOverrideResource) getDeviceByID(ctx context.Context, id string) (*unifi.DeviceConfig, error) {
	devices, err := r.client.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	for _, d := range devices.NetworkDevices {
		if d.ID == id {
			return r.client.GetDeviceByMAC(ctx, d.MAC)
		}
	}

	return nil, fmt.Errorf("device with ID '%s' not found", id)
}

func (r *DevicePortOverrideResource) findPortOverride(overrides []unifi.PortOverride, portIdx int) *unifi.PortOverride {
	for i := range overrides {
		if overrides[i].PortIdx != nil && *overrides[i].PortIdx == portIdx {
			return &overrides[i]
		}
	}
	return nil
}

func (r *DevicePortOverrideResource) mergePortOverride(overrides []unifi.PortOverride, newOverride *unifi.PortOverride, portIdx int) []unifi.PortOverride {
	for i := range overrides {
		if overrides[i].PortIdx != nil && *overrides[i].PortIdx == portIdx {
			overrides[i] = *newOverride
			return overrides
		}
	}
	return append(overrides, *newOverride)
}

func (r *DevicePortOverrideResource) removePortOverride(overrides []unifi.PortOverride, portIdx int) []unifi.PortOverride {
	result := make([]unifi.PortOverride, 0, len(overrides))
	for _, o := range overrides {
		if o.PortIdx == nil || *o.PortIdx != portIdx {
			result = append(result, o)
		}
	}
	return result
}

func (r *DevicePortOverrideResource) planToSDK(ctx context.Context, plan *DevicePortOverrideResourceModel, diags *diag.Diagnostics) *unifi.PortOverride {
	portIdx := int(plan.PortIdx.ValueInt64())

	override := &unifi.PortOverride{
		PortIdx:           &portIdx,
		SettingPreference: "manual",
	}

	if !plan.Name.IsNull() {
		override.Name = plan.Name.ValueString()
	}

	if !plan.PortProfileID.IsNull() {
		override.PortconfID = plan.PortProfileID.ValueString()
	}

	if !plan.PoeMode.IsNull() {
		override.PoeMode = plan.PoeMode.ValueString()
	}

	if !plan.OpMode.IsNull() {
		override.OpMode = plan.OpMode.ValueString()
	}

	if !plan.AggregateMembers.IsNull() {
		var members []int64
		diags.Append(plan.AggregateMembers.ElementsAs(ctx, &members, false)...)
		if diags.HasError() {
			return nil
		}
		intMembers := make([]int, len(members))
		for i, m := range members {
			intMembers[i] = int(m)
		}
		override.AggregateMembers = intMembers
	}

	if !plan.NativeNetworkID.IsNull() {
		override.NativeNetworkconfID = plan.NativeNetworkID.ValueString()
	}

	if !plan.TaggedNetworkIDs.IsNull() {
		var ids []string
		diags.Append(plan.TaggedNetworkIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil
		}
		override.TaggedNetworkconfIDs = ids
	}

	if !plan.ExcludedNetworkIDs.IsNull() {
		var ids []string
		diags.Append(plan.ExcludedNetworkIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil
		}
		override.ExcludedNetworkconfIDs = ids
	}

	if !plan.VoiceNetworkID.IsNull() {
		override.VoiceNetworkconfID = plan.VoiceNetworkID.ValueString()
	}

	if !plan.Autoneg.IsNull() {
		override.Autoneg = boolPtr(plan.Autoneg.ValueBool())
	}

	if !plan.Speed.IsNull() {
		override.Speed = intPtr(plan.Speed.ValueInt64())
	}

	if !plan.FullDuplex.IsNull() {
		override.FullDuplex = boolPtr(plan.FullDuplex.ValueBool())
	}

	if !plan.Isolation.IsNull() {
		override.Isolation = boolPtr(plan.Isolation.ValueBool())
	}

	if !plan.StpPortMode.IsNull() {
		override.StpPortMode = boolPtr(plan.StpPortMode.ValueBool())
	}

	if !plan.EgressRateLimitKbps.IsNull() {
		limit := int(plan.EgressRateLimitKbps.ValueInt64())
		if limit > 0 {
			override.EgressRateLimitKbpsEnabled = boolPtr(true)
			override.EgressRateLimitKbps = &limit
		} else {
			override.EgressRateLimitKbpsEnabled = boolPtr(false)
		}
	}

	if !plan.PortSecurityEnabled.IsNull() {
		override.PortSecurityEnabled = boolPtr(plan.PortSecurityEnabled.ValueBool())
	}

	if !plan.PortSecurityMacAddresses.IsNull() {
		var macs []string
		diags.Append(plan.PortSecurityMacAddresses.ElementsAs(ctx, &macs, false)...)
		if diags.HasError() {
			return nil
		}
		override.PortSecurityMacAddress = macs
	}

	return override
}

func (r *DevicePortOverrideResource) sdkToState(ctx context.Context, override *unifi.PortOverride, state *DevicePortOverrideResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Name = stringValueOrNull(override.Name)
	state.PortProfileID = stringValueOrNull(override.PortconfID)
	state.PoeMode = stringValueOrNull(override.PoeMode)
	state.OpMode = stringValueOrNull(override.OpMode)
	state.NativeNetworkID = stringValueOrNull(override.NativeNetworkconfID)
	state.VoiceNetworkID = stringValueOrNull(override.VoiceNetworkconfID)

	if len(override.AggregateMembers) > 0 {
		members := make([]int64, len(override.AggregateMembers))
		for i, m := range override.AggregateMembers {
			members[i] = int64(m)
		}
		memberSet, d := types.SetValueFrom(ctx, types.Int64Type, members)
		diags.Append(d...)
		state.AggregateMembers = memberSet
	} else {
		state.AggregateMembers = types.SetNull(types.Int64Type)
	}

	if len(override.TaggedNetworkconfIDs) > 0 {
		taggedSet, d := types.SetValueFrom(ctx, types.StringType, override.TaggedNetworkconfIDs)
		diags.Append(d...)
		state.TaggedNetworkIDs = taggedSet
	} else {
		state.TaggedNetworkIDs = types.SetNull(types.StringType)
	}

	if len(override.ExcludedNetworkconfIDs) > 0 {
		excludedSet, d := types.SetValueFrom(ctx, types.StringType, override.ExcludedNetworkconfIDs)
		diags.Append(d...)
		state.ExcludedNetworkIDs = excludedSet
	} else {
		state.ExcludedNetworkIDs = types.SetNull(types.StringType)
	}

	if override.Autoneg != nil {
		state.Autoneg = types.BoolValue(*override.Autoneg)
	} else {
		state.Autoneg = types.BoolNull()
	}

	if override.Speed != nil {
		state.Speed = types.Int64Value(int64(*override.Speed))
	} else {
		state.Speed = types.Int64Null()
	}

	if override.FullDuplex != nil {
		state.FullDuplex = types.BoolValue(*override.FullDuplex)
	} else {
		state.FullDuplex = types.BoolNull()
	}

	if override.Isolation != nil {
		state.Isolation = types.BoolValue(*override.Isolation)
	} else {
		state.Isolation = types.BoolNull()
	}

	if override.StpPortMode != nil {
		state.StpPortMode = types.BoolValue(*override.StpPortMode)
	} else {
		state.StpPortMode = types.BoolNull()
	}

	if override.EgressRateLimitKbps != nil && *override.EgressRateLimitKbps > 0 {
		state.EgressRateLimitKbps = types.Int64Value(int64(*override.EgressRateLimitKbps))
	} else {
		state.EgressRateLimitKbps = types.Int64Null()
	}

	if override.PortSecurityEnabled != nil {
		state.PortSecurityEnabled = types.BoolValue(*override.PortSecurityEnabled)
	} else {
		state.PortSecurityEnabled = types.BoolNull()
	}

	if len(override.PortSecurityMacAddress) > 0 {
		macSet, d := types.SetValueFrom(ctx, types.StringType, override.PortSecurityMacAddress)
		diags.Append(d...)
		state.PortSecurityMacAddresses = macSet
	} else {
		state.PortSecurityMacAddresses = types.SetNull(types.StringType)
	}

	return diags
}
