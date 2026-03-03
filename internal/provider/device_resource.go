package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &DeviceResource{}
	_ resource.ResourceWithImportState = &DeviceResource{}
)

type DeviceResource struct {
	client *AutoLoginClient
}

type DeviceResourceModel struct {
	ID                         types.String   `tfsdk:"id"`
	MAC                        types.String   `tfsdk:"mac"`
	Name                       types.String   `tfsdk:"name"`
	LedOverride                types.String   `tfsdk:"led_override"`
	LedOverrideColor           types.String   `tfsdk:"led_override_color"`
	LedOverrideColorBrightness types.Int64    `tfsdk:"led_override_color_brightness"`
	SNMPContact                types.String   `tfsdk:"snmp_contact"`
	SNMPLocation               types.String   `tfsdk:"snmp_location"`
	RadioOverrides             types.List     `tfsdk:"radio_overrides"`
	SiteID                     types.String   `tfsdk:"site_id"`
	Model                      types.String   `tfsdk:"model"`
	Type                       types.String   `tfsdk:"type"`
	IP                         types.String   `tfsdk:"ip"`
	Version                    types.String   `tfsdk:"version"`
	State                      types.Int64    `tfsdk:"state"`
	Adopted                    types.Bool     `tfsdk:"adopted"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
}

type RadioOverrideModel struct {
	Radio                 types.String `tfsdk:"radio"`
	Name                  types.String `tfsdk:"name"`
	Channel               types.Int64  `tfsdk:"channel"`
	ChannelWidth          types.Int64  `tfsdk:"channel_width"`
	TxPowerMode           types.String `tfsdk:"tx_power_mode"`
	TxPower               types.Int64  `tfsdk:"tx_power"`
	MinRSSIEnabled        types.Bool   `tfsdk:"min_rssi_enabled"`
	MinRSSI               types.Int64  `tfsdk:"min_rssi"`
	AntennaGain           types.Int64  `tfsdk:"antenna_gain"`
	VWireEnabled          types.Bool   `tfsdk:"vwire_enabled"`
	LoadBalanceEnabled    types.Bool   `tfsdk:"load_balance_enabled"`
	HardNoiseFloorEnabled types.Bool   `tfsdk:"hard_noise_floor_enabled"`
	SensLevelEnabled      types.Bool   `tfsdk:"sens_level_enabled"`
}

var radioOverrideAttrTypes = map[string]attr.Type{
	"radio":                    types.StringType,
	"name":                     types.StringType,
	"channel":                  types.Int64Type,
	"channel_width":            types.Int64Type,
	"tx_power_mode":            types.StringType,
	"tx_power":                 types.Int64Type,
	"min_rssi_enabled":         types.BoolType,
	"min_rssi":                 types.Int64Type,
	"antenna_gain":             types.Int64Type,
	"vwire_enabled":            types.BoolType,
	"load_balance_enabled":     types.BoolType,
	"hard_noise_floor_enabled": types.BoolType,
	"sens_level_enabled":       types.BoolType,
}

func NewDeviceResource() resource.Resource {
	return &DeviceResource{}
}

func (r *DeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *DeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi device's configuration. Devices are physically adopted hardware — " +
			"this resource manages writable settings (name, LED, SNMP) on an existing device. " +
			"Create looks up the device by MAC; delete removes it from Terraform state without forgetting the device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the device.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mac": schema.StringAttribute{
				Description: "The MAC address of the device. Used to identify the device on create.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the device.",
				Optional:    true,
				Computed:    true,
			},
			"led_override": schema.StringAttribute{
				Description: "LED override mode. Valid values: 'default', 'on', 'off'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("default", "on", "off"),
				},
			},
			"led_override_color": schema.StringAttribute{
				Description: "LED override color (hex).",
				Optional:    true,
				Computed:    true,
			},
			"led_override_color_brightness": schema.Int64Attribute{
				Description: "LED override color brightness (0-100).",
				Optional:    true,
				Computed:    true,
			},
			"snmp_contact": schema.StringAttribute{
				Description: "SNMP contact string.",
				Optional:    true,
				Computed:    true,
			},
			"snmp_location": schema.StringAttribute{
				Description: "SNMP location string.",
				Optional:    true,
				Computed:    true,
			},
			"radio_overrides": schema.ListNestedAttribute{
				Description: "Radio configuration overrides for access points.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"radio": schema.StringAttribute{
							Description: "Radio band. Valid values: 'ng' (2.4GHz), 'na' (5GHz), '6e' (6GHz).",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("ng", "na", "6e"),
							},
						},
						"name": schema.StringAttribute{
							Description: "Radio name.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"channel": schema.Int64Attribute{
							Description: "Radio channel.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"channel_width": schema.Int64Attribute{
							Description: "Channel width (HT mode).",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"tx_power_mode": schema.StringAttribute{
							Description: "Transmit power mode. Valid values: 'auto', 'low', 'medium', 'high', 'custom'.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("auto", "low", "medium", "high", "custom"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"tx_power": schema.Int64Attribute{
							Description: "Custom transmit power (dBm).",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"min_rssi_enabled": schema.BoolAttribute{
							Description: "Enable minimum RSSI.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"min_rssi": schema.Int64Attribute{
							Description: "Minimum RSSI value.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"antenna_gain": schema.Int64Attribute{
							Description: "Antenna gain (dBi).",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"vwire_enabled": schema.BoolAttribute{
							Description: "Enable virtual wire.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"load_balance_enabled": schema.BoolAttribute{
							Description: "Enable load balancing.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"hard_noise_floor_enabled": schema.BoolAttribute{
							Description: "Enable hard noise floor.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"sens_level_enabled": schema.BoolAttribute{
							Description: "Enable sensitivity level.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the device is adopted.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The device model identifier.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The device type (uap, usw, ugw, uxg, udm).",
				Computed:    true,
			},
			"ip": schema.StringAttribute{
				Description: "The IP address of the device.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The firmware version of the device.",
				Computed:    true,
			},
			"state": schema.Int64Attribute{
				Description: "The device state (0=offline, 1=connected, 2=pending adoption).",
				Computed:    true,
			},
			"adopted": schema.BoolAttribute{
				Description: "Whether the device is adopted.",
				Computed:    true,
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

func (r *DeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceResourceModel

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

	mac := plan.MAC.ValueString()
	device, err := r.client.GetDeviceByMAC(ctx, mac)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "find", "device")
		return
	}

	updateDevice := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updateDevice.MAC = mac
	updated, err := r.client.UpdateDevice(ctx, device.ID, updateDevice)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "device")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceResourceModel

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

	device, err := r.client.GetDeviceByMAC(ctx, state.MAC.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "device")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, device, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceResourceModel

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

	var state DeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateDevice := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updateDevice.MAC = plan.MAC.ValueString()

	updated, err := r.client.UpdateDevice(ctx, state.ID.ValueString(), updateDevice)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "device")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Devices are physical hardware — delete just removes from Terraform state.
	// The device remains adopted on the controller.
}

func (r *DeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	mac := req.ID

	device, err := r.client.GetDeviceByMAC(ctx, mac)
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Error",
			fmt.Sprintf("Could not find device with MAC %q: %s. Import requires the device MAC address (e.g., aa:bb:cc:dd:ee:ff).", mac, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), device.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mac"), device.MAC)...)
}

func (r *DeviceResource) planToSDK(ctx context.Context, plan *DeviceResourceModel, diags *diag.Diagnostics) *unifi.DeviceConfig {
	device := &unifi.DeviceConfig{}

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		device.Name = plan.Name.ValueString()
	}
	if !plan.LedOverride.IsNull() && !plan.LedOverride.IsUnknown() {
		device.LedOverride = plan.LedOverride.ValueString()
	}
	if !plan.LedOverrideColor.IsNull() && !plan.LedOverrideColor.IsUnknown() {
		device.LedOverrideColor = plan.LedOverrideColor.ValueString()
	}
	if !plan.LedOverrideColorBrightness.IsNull() && !plan.LedOverrideColorBrightness.IsUnknown() {
		v := int(plan.LedOverrideColorBrightness.ValueInt64())
		device.LedOverrideColorBrightness = &v
	}
	if !plan.SNMPContact.IsNull() && !plan.SNMPContact.IsUnknown() {
		device.SNMPContact = plan.SNMPContact.ValueString()
	}
	if !plan.SNMPLocation.IsNull() && !plan.SNMPLocation.IsUnknown() {
		device.SNMPLocation = plan.SNMPLocation.ValueString()
	}

	if !plan.RadioOverrides.IsNull() && !plan.RadioOverrides.IsUnknown() {
		var overrides []RadioOverrideModel
		diags.Append(plan.RadioOverrides.ElementsAs(ctx, &overrides, false)...)
		for _, o := range overrides {
			ro := unifi.RadioOverride{
				Radio: o.Radio.ValueString(),
			}
			if !o.Name.IsNull() && !o.Name.IsUnknown() {
				ro.Name = o.Name.ValueString()
			}
			if !o.Channel.IsNull() && !o.Channel.IsUnknown() {
				ro.Channel = intPtr(o.Channel.ValueInt64())
			}
			if !o.ChannelWidth.IsNull() && !o.ChannelWidth.IsUnknown() {
				ro.ChannelWidth = intPtr(o.ChannelWidth.ValueInt64())
			}
			if !o.TxPowerMode.IsNull() && !o.TxPowerMode.IsUnknown() {
				ro.TxPowerMode = o.TxPowerMode.ValueString()
			}
			if !o.TxPower.IsNull() && !o.TxPower.IsUnknown() {
				ro.TxPower = intPtr(o.TxPower.ValueInt64())
			}
			if !o.MinRSSIEnabled.IsNull() && !o.MinRSSIEnabled.IsUnknown() {
				ro.MinRSSIEnabled = boolPtr(o.MinRSSIEnabled.ValueBool())
			}
			if !o.MinRSSI.IsNull() && !o.MinRSSI.IsUnknown() {
				ro.MinRSSI = intPtr(o.MinRSSI.ValueInt64())
			}
			if !o.AntennaGain.IsNull() && !o.AntennaGain.IsUnknown() {
				ro.AntennaGain = intPtr(o.AntennaGain.ValueInt64())
			}
			if !o.VWireEnabled.IsNull() && !o.VWireEnabled.IsUnknown() {
				ro.VWireEnabled = boolPtr(o.VWireEnabled.ValueBool())
			}
			if !o.LoadBalanceEnabled.IsNull() && !o.LoadBalanceEnabled.IsUnknown() {
				ro.LoadBalanceEnable = boolPtr(o.LoadBalanceEnabled.ValueBool())
			}
			if !o.HardNoiseFloorEnabled.IsNull() && !o.HardNoiseFloorEnabled.IsUnknown() {
				ro.HardNoiseFloor = boolPtr(o.HardNoiseFloorEnabled.ValueBool())
			}
			if !o.SensLevelEnabled.IsNull() && !o.SensLevelEnabled.IsUnknown() {
				ro.SensLevelEnabled = boolPtr(o.SensLevelEnabled.ValueBool())
			}
			device.RadioOverrides = append(device.RadioOverrides, ro)
		}
	}

	return device
}

func (r *DeviceResource) sdkToState(ctx context.Context, device *unifi.DeviceConfig, state *DeviceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(device.ID)
	state.MAC = types.StringValue(device.MAC)
	state.SiteID = types.StringValue(device.SiteID)
	state.Name = stringValueOrNull(device.Name)
	state.Model = stringValueOrNull(device.Model)
	state.Type = stringValueOrNull(device.Type)
	state.IP = stringValueOrNull(device.IP)
	state.Version = stringValueOrNull(device.Version)
	state.LedOverride = stringValueOrNull(device.LedOverride)
	state.LedOverrideColor = stringValueOrNull(device.LedOverrideColor)
	state.SNMPContact = stringValueOrNull(device.SNMPContact)
	state.SNMPLocation = stringValueOrNull(device.SNMPLocation)

	if device.LedOverrideColorBrightness != nil {
		state.LedOverrideColorBrightness = types.Int64Value(int64(*device.LedOverrideColorBrightness))
	} else {
		state.LedOverrideColorBrightness = types.Int64Null()
	}

	if device.State != nil {
		state.State = types.Int64Value(int64(*device.State))
	} else {
		state.State = types.Int64Null()
	}

	state.Adopted = types.BoolValue(derefBool(device.Adopted))

	if len(device.RadioOverrides) > 0 {
		overrideValues := make([]attr.Value, len(device.RadioOverrides))
		for i, ro := range device.RadioOverrides {
			vals := map[string]attr.Value{
				"radio":                    stringValueOrNull(ro.Radio),
				"name":                     stringValueOrNull(ro.Name),
				"tx_power_mode":            stringValueOrNull(ro.TxPowerMode),
				"min_rssi_enabled":         types.BoolValue(derefBool(ro.MinRSSIEnabled)),
				"vwire_enabled":            types.BoolValue(derefBool(ro.VWireEnabled)),
				"load_balance_enabled":     types.BoolValue(derefBool(ro.LoadBalanceEnable)),
				"hard_noise_floor_enabled": types.BoolValue(derefBool(ro.HardNoiseFloor)),
				"sens_level_enabled":       types.BoolValue(derefBool(ro.SensLevelEnabled)),
			}
			if ro.Channel != nil {
				vals["channel"] = types.Int64Value(int64(*ro.Channel))
			} else {
				vals["channel"] = types.Int64Null()
			}
			if ro.ChannelWidth != nil {
				vals["channel_width"] = types.Int64Value(int64(*ro.ChannelWidth))
			} else {
				vals["channel_width"] = types.Int64Null()
			}
			if ro.TxPower != nil {
				vals["tx_power"] = types.Int64Value(int64(*ro.TxPower))
			} else {
				vals["tx_power"] = types.Int64Null()
			}
			if ro.MinRSSI != nil {
				vals["min_rssi"] = types.Int64Value(int64(*ro.MinRSSI))
			} else {
				vals["min_rssi"] = types.Int64Null()
			}
			if ro.AntennaGain != nil {
				vals["antenna_gain"] = types.Int64Value(int64(*ro.AntennaGain))
			} else {
				vals["antenna_gain"] = types.Int64Null()
			}

			obj, d := types.ObjectValue(radioOverrideAttrTypes, vals)
			diags.Append(d...)
			overrideValues[i] = obj
		}
		list, d := types.ListValue(types.ObjectType{AttrTypes: radioOverrideAttrTypes}, overrideValues)
		diags.Append(d...)
		state.RadioOverrides = list
	} else {
		state.RadioOverrides = types.ListValueMust(types.ObjectType{AttrTypes: radioOverrideAttrTypes}, []attr.Value{})
	}

	return diags
}
