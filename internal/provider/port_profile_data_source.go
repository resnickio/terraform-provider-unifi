package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ datasource.DataSource = &PortProfileDataSource{}

type PortProfileDataSource struct {
	client *AutoLoginClient
}

type PortProfileDataSourceModel struct {
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
}

func NewPortProfileDataSource() datasource.DataSource {
	return &PortProfileDataSource{}
}

func (d *PortProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_profile"
}

func (d *PortProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi switch port profile. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the port profile. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the port profile. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the port profile exists.",
				Computed:    true,
			},
			"native_network_id": schema.StringAttribute{
				Description: "Network ID for the native/untagged VLAN.",
				Computed:    true,
			},
			"tagged_vlan_mgmt": schema.StringAttribute{
				Description: "Tagged VLAN management mode (all, block, custom).",
				Computed:    true,
			},
			"excluded_network_ids": schema.SetAttribute{
				Description: "Set of network IDs to exclude from tagged VLANs.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"autoneg": schema.BoolAttribute{
				Description: "Enable auto-negotiation for link speed and duplex.",
				Computed:    true,
			},
			"speed": schema.Int64Attribute{
				Description: "Link speed in Mbps.",
				Computed:    true,
			},
			"full_duplex": schema.BoolAttribute{
				Description: "Enable full duplex mode.",
				Computed:    true,
			},
			"poe_mode": schema.StringAttribute{
				Description: "PoE mode (auto, pasv24, passthrough, off).",
				Computed:    true,
			},
			"op_mode": schema.StringAttribute{
				Description: "Port operation mode (switch, mirror, aggregate).",
				Computed:    true,
			},
			"isolation": schema.BoolAttribute{
				Description: "Enable port isolation.",
				Computed:    true,
			},
			"dot1x_ctrl": schema.StringAttribute{
				Description: "802.1X control mode.",
				Computed:    true,
			},
			"dot1x_idle_timeout": schema.Int64Attribute{
				Description: "802.1X idle timeout in seconds.",
				Computed:    true,
			},
			"stp_port_mode": schema.BoolAttribute{
				Description: "Enable STP on this port.",
				Computed:    true,
			},
			"lldpmed_enabled": schema.BoolAttribute{
				Description: "Enable LLDP-MED.",
				Computed:    true,
			},
			"lldpmed_notify_enabled": schema.BoolAttribute{
				Description: "Enable LLDP-MED topology change notifications.",
				Computed:    true,
			},
			"stormctrl_bcast_enabled": schema.BoolAttribute{
				Description: "Enable broadcast storm control.",
				Computed:    true,
			},
			"stormctrl_bcast_rate": schema.Int64Attribute{
				Description: "Broadcast storm control rate limit.",
				Computed:    true,
			},
			"stormctrl_mcast_enabled": schema.BoolAttribute{
				Description: "Enable multicast storm control.",
				Computed:    true,
			},
			"stormctrl_mcast_rate": schema.Int64Attribute{
				Description: "Multicast storm control rate limit.",
				Computed:    true,
			},
			"stormctrl_ucast_enabled": schema.BoolAttribute{
				Description: "Enable unicast storm control.",
				Computed:    true,
			},
			"stormctrl_ucast_rate": schema.Int64Attribute{
				Description: "Unicast storm control rate limit.",
				Computed:    true,
			},
			"egress_rate_limit_kbps_enabled": schema.BoolAttribute{
				Description: "Enable egress rate limiting.",
				Computed:    true,
			},
			"egress_rate_limit_kbps": schema.Int64Attribute{
				Description: "Egress rate limit in Kbps.",
				Computed:    true,
			},
			"port_security_enabled": schema.BoolAttribute{
				Description: "Enable port security (MAC address filtering).",
				Computed:    true,
			},
			"port_security_mac_address": schema.SetAttribute{
				Description: "Set of allowed MAC addresses when port security is enabled.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"port_keepalive_enabled": schema.BoolAttribute{
				Description: "Enable port keepalive.",
				Computed:    true,
			},
		},
	}
}

func (d *PortProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PortProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PortProfileDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a port profile.",
		)
		return
	}

	var profile *unifi.PortConf
	var err error

	if hasID {
		profile, err = d.client.GetPortProfile(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "port profile")
			return
		}
	} else {
		profiles, err := d.client.ListPortProfiles(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "port profiles")
			return
		}

		searchName := config.Name.ValueString()
		for i := range profiles {
			if profiles[i].Name == searchName {
				profile = &profiles[i]
				break
			}
		}

		if profile == nil {
			resp.Diagnostics.AddError(
				"Port Profile Not Found",
				fmt.Sprintf("No port profile found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, profile, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *PortProfileDataSource) sdkToState(ctx context.Context, profile *unifi.PortConf, state *PortProfileDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(profile.ID)
	state.SiteID = types.StringValue(profile.SiteID)
	state.Name = types.StringValue(profile.Name)

	state.NativeNetworkID = stringValueOrNull(profile.NativeNetworkconfID)
	state.TaggedVlanMgmt = stringValueOrNull(profile.TaggedVlanMgmt)

	if profile.ExcludedNetworkconfIDs == nil {
		profile.ExcludedNetworkconfIDs = []string{}
	}
	excludedIDs, diagExcl := types.SetValueFrom(ctx, types.StringType, profile.ExcludedNetworkconfIDs)
	diags.Append(diagExcl...)
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
	macs, diagMacs := types.SetValueFrom(ctx, types.StringType, profile.PortSecurityMacAddress)
	diags.Append(diagMacs...)
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
