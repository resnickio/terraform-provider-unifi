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

var _ datasource.DataSource = &NetworkDataSource{}

type NetworkDataSource struct {
	client *AutoLoginClient
}

type NetworkDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	SiteID  types.String `tfsdk:"site_id"`
	Purpose types.String `tfsdk:"purpose"`
	Enabled types.Bool   `tfsdk:"enabled"`

	// VLAN
	VlanID types.Int64  `tfsdk:"vlan_id"`
	Subnet types.String `tfsdk:"subnet"`

	// DHCP Core
	DHCPEnabled types.Bool   `tfsdk:"dhcp_enabled"`
	DHCPStart   types.String `tfsdk:"dhcp_start"`
	DHCPStop    types.String `tfsdk:"dhcp_stop"`
	DHCPLease   types.Int64  `tfsdk:"dhcp_lease"`

	// DHCP DNS
	DHCPDNS types.Set `tfsdk:"dhcp_dns"`

	// DHCP Gateway
	DHCPGatewayEnabled types.Bool   `tfsdk:"dhcp_gateway_enabled"`
	DHCPGateway        types.String `tfsdk:"dhcp_gateway"`

	// DHCP NTP
	DHCPNTPEnabled types.Bool `tfsdk:"dhcp_ntp_enabled"`
	DHCPNTP        types.Set  `tfsdk:"dhcp_ntp"`

	// DHCP Boot/PXE
	DHCPBootEnabled  types.Bool   `tfsdk:"dhcp_boot_enabled"`
	DHCPBootServer   types.String `tfsdk:"dhcp_boot_server"`
	DHCPBootFilename types.String `tfsdk:"dhcp_boot_filename"`

	// DHCP Additional Options
	DHCPRelayEnabled      types.Bool   `tfsdk:"dhcp_relay_enabled"`
	DHCPTimeOffsetEnabled types.Bool   `tfsdk:"dhcp_time_offset_enabled"`
	DHCPUnifiController   types.String `tfsdk:"dhcp_unifi_controller"`
	DHCPWPADUrl           types.String `tfsdk:"dhcp_wpad_url"`
	DHCPGuardingEnabled   types.Bool   `tfsdk:"dhcp_guarding_enabled"`

	// Multicast
	DomainName        types.String `tfsdk:"domain_name"`
	IGMPSnooping      types.Bool   `tfsdk:"igmp_snooping"`
	IGMPProxyUpstream types.Bool   `tfsdk:"igmp_proxy_upstream"`

	// Network Access
	InternetAccessEnabled     types.Bool `tfsdk:"internet_access_enabled"`
	IntraNetworkAccessEnabled types.Bool `tfsdk:"intra_network_access_enabled"`
	NATEnabled                types.Bool `tfsdk:"nat_enabled"`
	MDNSEnabled               types.Bool `tfsdk:"mdns_enabled"`
	UPnPLANEnabled            types.Bool `tfsdk:"upnp_lan_enabled"`

	// Routing
	NetworkGroup   types.String `tfsdk:"network_group"`
	FirewallZoneID types.String `tfsdk:"firewall_zone_id"`

	// IPv6
	IPv6SettingPreference types.String `tfsdk:"ipv6_setting_preference"`
}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

func (d *NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi network. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the network. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the network. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the network exists.",
				Computed:    true,
			},
			"purpose": schema.StringAttribute{
				Description: "The purpose of the network (corporate, guest, wan, vlan-only).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the network is enabled.",
				Computed:    true,
			},

			// VLAN
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID for this network.",
				Computed:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "The subnet in CIDR notation.",
				Computed:    true,
			},

			// DHCP Core
			"dhcp_enabled": schema.BoolAttribute{
				Description: "Whether DHCP is enabled on this network.",
				Computed:    true,
			},
			"dhcp_start": schema.StringAttribute{
				Description: "The start of the DHCP IP range.",
				Computed:    true,
			},
			"dhcp_stop": schema.StringAttribute{
				Description: "The end of the DHCP IP range.",
				Computed:    true,
			},
			"dhcp_lease": schema.Int64Attribute{
				Description: "The DHCP lease time in seconds.",
				Computed:    true,
			},

			// DHCP DNS
			"dhcp_dns": schema.SetAttribute{
				Description: "Set of DNS servers provided via DHCP.",
				Computed:    true,
				ElementType: types.StringType,
			},

			// DHCP Gateway
			"dhcp_gateway_enabled": schema.BoolAttribute{
				Description: "Whether a custom gateway is provided via DHCP (Option 3).",
				Computed:    true,
			},
			"dhcp_gateway": schema.StringAttribute{
				Description: "Custom gateway IP address provided via DHCP (Option 3).",
				Computed:    true,
			},

			// DHCP NTP
			"dhcp_ntp_enabled": schema.BoolAttribute{
				Description: "Whether NTP servers are provided via DHCP (Option 42).",
				Computed:    true,
			},
			"dhcp_ntp": schema.SetAttribute{
				Description: "Set of NTP servers provided via DHCP.",
				Computed:    true,
				ElementType: types.StringType,
			},

			// DHCP Boot/PXE
			"dhcp_boot_enabled": schema.BoolAttribute{
				Description: "Whether DHCP network boot (PXE) is enabled.",
				Computed:    true,
			},
			"dhcp_boot_server": schema.StringAttribute{
				Description: "The IP address of the boot server (DHCP Option 66).",
				Computed:    true,
			},
			"dhcp_boot_filename": schema.StringAttribute{
				Description: "The boot filename provided to clients (DHCP Option 67).",
				Computed:    true,
			},

			// DHCP Additional Options
			"dhcp_relay_enabled": schema.BoolAttribute{
				Description: "Whether DHCP relay is enabled.",
				Computed:    true,
			},
			"dhcp_time_offset_enabled": schema.BoolAttribute{
				Description: "Whether time offset is provided via DHCP (Option 2).",
				Computed:    true,
			},
			"dhcp_unifi_controller": schema.StringAttribute{
				Description: "UniFi controller IP address provided via DHCP (Option 43).",
				Computed:    true,
			},
			"dhcp_wpad_url": schema.StringAttribute{
				Description: "Web Proxy Auto-Discovery (WPAD) URL provided via DHCP (Option 252).",
				Computed:    true,
			},
			"dhcp_guarding_enabled": schema.BoolAttribute{
				Description: "Whether DHCP guarding is enabled.",
				Computed:    true,
			},

			// Multicast
			"domain_name": schema.StringAttribute{
				Description: "The domain name for this network.",
				Computed:    true,
			},
			"igmp_snooping": schema.BoolAttribute{
				Description: "Whether IGMP snooping is enabled.",
				Computed:    true,
			},
			"igmp_proxy_upstream": schema.BoolAttribute{
				Description: "Whether this network acts as an IGMP proxy upstream interface.",
				Computed:    true,
			},

			// Network Access
			"internet_access_enabled": schema.BoolAttribute{
				Description: "Whether internet access is enabled for this network.",
				Computed:    true,
			},
			"intra_network_access_enabled": schema.BoolAttribute{
				Description: "Whether devices on this network can communicate with devices on other networks.",
				Computed:    true,
			},
			"nat_enabled": schema.BoolAttribute{
				Description: "Whether NAT is enabled for this network.",
				Computed:    true,
			},
			"mdns_enabled": schema.BoolAttribute{
				Description: "Whether mDNS (Bonjour/Avahi) is enabled for this network.",
				Computed:    true,
			},
			"upnp_lan_enabled": schema.BoolAttribute{
				Description: "Whether UPnP is enabled on this LAN network.",
				Computed:    true,
			},

			// Routing
			"network_group": schema.StringAttribute{
				Description: "The network group (LAN, WAN, WAN2).",
				Computed:    true,
			},
			"firewall_zone_id": schema.StringAttribute{
				Description: "The firewall zone ID associated with this network.",
				Computed:    true,
			},

			// IPv6
			"ipv6_setting_preference": schema.StringAttribute{
				Description: "IPv6 configuration preference (auto, manual).",
				Computed:    true,
			},
		},
	}
}

func (d *NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NetworkDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a network.",
		)
		return
	}

	var network *unifi.Network
	var err error

	if hasID {
		network, err = d.client.GetNetwork(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "network")
			return
		}
	} else {
		networks, err := d.client.ListNetworks(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "networks")
			return
		}

		searchName := config.Name.ValueString()
		for i := range networks {
			if networks[i].Name == searchName {
				network = &networks[i]
				break
			}
		}

		if network == nil {
			resp.Diagnostics.AddError(
				"Network Not Found",
				fmt.Sprintf("No network found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, network, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *NetworkDataSource) sdkToState(ctx context.Context, network *unifi.Network, state *NetworkDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(network.ID)
	state.SiteID = types.StringValue(network.SiteID)
	state.Name = types.StringValue(network.Name)
	state.Purpose = types.StringValue(network.Purpose)
	state.Enabled = types.BoolValue(derefBool(network.Enabled))

	// VLAN
	if network.VLAN != nil {
		state.VlanID = types.Int64Value(int64(*network.VLAN))
	} else {
		state.VlanID = types.Int64Null()
	}
	state.Subnet = stringValueOrNull(network.IPSubnet)

	// DHCP Core
	state.DHCPEnabled = types.BoolValue(derefBool(network.DHCPDEnabled))
	state.DHCPStart = stringValueOrNull(network.DHCPDStart)
	state.DHCPStop = stringValueOrNull(network.DHCPDStop)
	if network.DHCPDLeasetime != nil {
		state.DHCPLease = types.Int64Value(int64(*network.DHCPDLeasetime))
	} else {
		state.DHCPLease = types.Int64Null()
	}

	// DHCP DNS
	var dnsServers []string
	if network.DHCPDDns1 != "" {
		dnsServers = append(dnsServers, network.DHCPDDns1)
	}
	if network.DHCPDDns2 != "" {
		dnsServers = append(dnsServers, network.DHCPDDns2)
	}
	if network.DHCPDDns3 != "" {
		dnsServers = append(dnsServers, network.DHCPDDns3)
	}
	if network.DHCPDDns4 != "" {
		dnsServers = append(dnsServers, network.DHCPDDns4)
	}
	if len(dnsServers) > 0 {
		dnsSet, d := types.SetValueFrom(ctx, types.StringType, dnsServers)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.DHCPDNS = dnsSet
	} else {
		state.DHCPDNS = types.SetNull(types.StringType)
	}

	// DHCP Gateway
	if network.DHCPDGatewayEnabled != nil {
		state.DHCPGatewayEnabled = types.BoolValue(*network.DHCPDGatewayEnabled)
	} else {
		state.DHCPGatewayEnabled = types.BoolNull()
	}
	state.DHCPGateway = stringValueOrNull(network.DHCPDGateway)

	// DHCP NTP
	if network.DHCPDNTPEnabled != nil {
		state.DHCPNTPEnabled = types.BoolValue(*network.DHCPDNTPEnabled)
	} else {
		state.DHCPNTPEnabled = types.BoolNull()
	}
	var ntpServers []string
	if network.DHCPDNtp1 != "" {
		ntpServers = append(ntpServers, network.DHCPDNtp1)
	}
	if network.DHCPDNtp2 != "" {
		ntpServers = append(ntpServers, network.DHCPDNtp2)
	}
	if len(ntpServers) > 0 {
		ntpSet, d := types.SetValueFrom(ctx, types.StringType, ntpServers)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.DHCPNTP = ntpSet
	} else {
		state.DHCPNTP = types.SetNull(types.StringType)
	}

	// DHCP Boot/PXE
	if network.DHCPDBootEnabled != nil {
		state.DHCPBootEnabled = types.BoolValue(*network.DHCPDBootEnabled)
	} else {
		state.DHCPBootEnabled = types.BoolNull()
	}
	state.DHCPBootServer = stringValueOrNull(network.DHCPDBootServer)
	state.DHCPBootFilename = stringValueOrNull(network.DHCPDBootFilename)

	// DHCP Additional Options
	if network.DHCPRelayEnabled != nil {
		state.DHCPRelayEnabled = types.BoolValue(*network.DHCPRelayEnabled)
	} else {
		state.DHCPRelayEnabled = types.BoolNull()
	}
	if network.DHCPDTimeOffsetEnabled != nil {
		state.DHCPTimeOffsetEnabled = types.BoolValue(*network.DHCPDTimeOffsetEnabled)
	} else {
		state.DHCPTimeOffsetEnabled = types.BoolNull()
	}
	state.DHCPUnifiController = stringValueOrNull(network.DHCPDUnifiController)
	state.DHCPWPADUrl = stringValueOrNull(network.DHCPDWPADUrl)
	if network.DHCPGuardingEnabled != nil {
		state.DHCPGuardingEnabled = types.BoolValue(*network.DHCPGuardingEnabled)
	} else {
		state.DHCPGuardingEnabled = types.BoolNull()
	}

	// Multicast
	state.DomainName = stringValueOrNull(network.DomainName)
	state.IGMPSnooping = types.BoolValue(derefBool(network.IGMPSnooping))
	if network.IGMPProxyUpstream != nil {
		state.IGMPProxyUpstream = types.BoolValue(*network.IGMPProxyUpstream)
	} else {
		state.IGMPProxyUpstream = types.BoolNull()
	}

	// Network Access
	state.InternetAccessEnabled = types.BoolValue(derefBool(network.InternetAccessEnabled))
	state.IntraNetworkAccessEnabled = types.BoolValue(derefBool(network.IntraNetworkAccessEnabled))
	state.NATEnabled = types.BoolValue(derefBool(network.IsNAT))
	if network.MDNSEnabled != nil {
		state.MDNSEnabled = types.BoolValue(*network.MDNSEnabled)
	} else {
		state.MDNSEnabled = types.BoolNull()
	}
	if network.UpnpLANEnabled != nil {
		state.UPnPLANEnabled = types.BoolValue(*network.UpnpLANEnabled)
	} else {
		state.UPnPLANEnabled = types.BoolNull()
	}

	// Routing
	state.NetworkGroup = types.StringValue(network.NetworkGroup)
	state.FirewallZoneID = stringValueOrNull(network.FirewallZoneID)

	// IPv6
	state.IPv6SettingPreference = stringValueOrNull(network.IPv6SettingPreference)

	return diags
}
