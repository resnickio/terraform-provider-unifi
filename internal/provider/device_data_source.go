package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ datasource.DataSource = &DeviceDataSource{}

type DeviceDataSource struct {
	client *AutoLoginClient
}

type DeviceDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	SiteID  types.String `tfsdk:"site_id"`
	MAC     types.String `tfsdk:"mac"`
	Name    types.String `tfsdk:"name"`
	Model   types.String `tfsdk:"model"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
	IP      types.String `tfsdk:"ip"`
	Adopted types.Bool   `tfsdk:"adopted"`
	State   types.Int64  `tfsdk:"state"`
}

func NewDeviceDataSource() datasource.DataSource {
	return &DeviceDataSource{}
}

func (d *DeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *DeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi device (switch, access point, gateway). Lookup by MAC address or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the device.",
				Computed:    true,
			},
			"mac": schema.StringAttribute{
				Description: "The MAC address of the device. Specify either mac or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the device. Specify either mac or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the device exists.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "The model of the device (e.g., 'USW-Pro-24-PoE', 'U6-Pro').",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of device: 'usw' (switch), 'uap' (access point), 'ugw' (gateway), 'udm' (dream machine), 'uxg' (next-gen gateway).",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The firmware version of the device.",
				Computed:    true,
			},
			"ip": schema.StringAttribute{
				Description: "The IP address of the device.",
				Computed:    true,
			},
			"adopted": schema.BoolAttribute{
				Description: "Whether the device has been adopted.",
				Computed:    true,
			},
			"state": schema.Int64Attribute{
				Description: "The state of the device: 0=offline, 1=connected, 2=pending.",
				Computed:    true,
			},
		},
	}
}

func (d *DeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DeviceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasMAC := !config.MAC.IsNull() && config.MAC.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasMAC && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'mac' or 'name' must be specified to look up a device.",
		)
		return
	}

	var device *unifi.DeviceConfig
	var err error

	if hasMAC {
		mac := strings.ToLower(strings.ReplaceAll(config.MAC.ValueString(), "-", ":"))
		device, err = d.client.GetDeviceByMAC(ctx, mac)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "device")
			return
		}
	} else {
		devices, err := d.client.ListDevices(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "devices")
			return
		}

		searchName := config.Name.ValueString()
		for i := range devices.NetworkDevices {
			if devices.NetworkDevices[i].Name == searchName {
				mac := devices.NetworkDevices[i].MAC
				device, err = d.client.GetDeviceByMAC(ctx, mac)
				if err != nil {
					handleSDKError(&resp.Diagnostics, err, "read", "device")
					return
				}
				break
			}
		}

		if device == nil {
			resp.Diagnostics.AddError(
				"Device Not Found",
				fmt.Sprintf("No device found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, device, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *DeviceDataSource) sdkToState(ctx context.Context, device *unifi.DeviceConfig, state *DeviceDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(device.ID)
	state.SiteID = stringValueOrNull(device.SiteID)
	state.MAC = types.StringValue(device.MAC)
	state.Name = stringValueOrNull(device.Name)
	state.Model = stringValueOrNull(device.Model)
	state.Type = stringValueOrNull(device.Type)
	state.Version = stringValueOrNull(device.Version)
	state.IP = stringValueOrNull(device.IP)

	if device.Adopted != nil {
		state.Adopted = types.BoolValue(*device.Adopted)
	} else {
		state.Adopted = types.BoolNull()
	}

	if device.State != nil {
		state.State = types.Int64Value(int64(*device.State))
	} else {
		state.State = types.Int64Null()
	}

	return diags
}
