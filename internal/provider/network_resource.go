package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

const (
	defaultDHCPLease = 86400
)

var (
	_ resource.Resource                = &NetworkResource{}
	_ resource.ResourceWithImportState = &NetworkResource{}
)

var ipv6AttrTypes = map[string]attr.Type{
	"interface_type":             types.StringType,
	"setting_preference":         types.StringType,
	"wan_delegation_type":        types.StringType,
	"subnet":                     types.StringType,
	"client_address_assignment":  types.StringType,
	"pd_interface":               types.StringType,
	"pd_prefixid":                types.StringType,
	"pd_start":                   types.StringType,
	"pd_stop":                    types.StringType,
	"pd_auto_prefixid_enabled":   types.BoolType,
	"ra_enabled":                 types.BoolType,
	"ra_preferred_lifetime":      types.Int64Type,
	"ra_priority":                types.StringType,
	"ra_valid_lifetime":          types.Int64Type,
	"dhcpv6_enabled":             types.BoolType,
	"dhcpv6_start":               types.StringType,
	"dhcpv6_stop":                types.StringType,
	"dhcpv6_lease_time":          types.Int64Type,
	"dhcpv6_dns_auto":            types.BoolType,
	"dhcpv6_dns":                 types.ListType{ElemType: types.StringType},
	"dhcpv6_allow_slaac":         types.BoolType,
}

type NetworkResource struct {
	client *AutoLoginClient
}

type NetworkResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	SiteID   types.String   `tfsdk:"site_id"`
	Name     types.String   `tfsdk:"name"`
	Purpose  types.String   `tfsdk:"purpose"`
	Enabled  types.Bool     `tfsdk:"enabled"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`

	// VLAN
	VlanID  types.Int64  `tfsdk:"vlan_id"`
	Subnet  types.String `tfsdk:"subnet"`

	// DHCP Core
	DHCPEnabled types.Bool   `tfsdk:"dhcp_enabled"`
	DHCPStart   types.String `tfsdk:"dhcp_start"`
	DHCPStop    types.String `tfsdk:"dhcp_stop"`
	DHCPLease   types.Int64  `tfsdk:"dhcp_lease"`

	// DHCP DNS
	DHCPDNSEnabled types.Bool `tfsdk:"dhcp_dns_enabled"`
	DHCPDNS        types.Set  `tfsdk:"dhcp_dns"`

	// DHCP Gateway
	DHCPGatewayEnabled types.Bool   `tfsdk:"dhcp_gateway_enabled"`
	DHCPGateway        types.String `tfsdk:"dhcp_gateway"`

	// DHCP NTP
	DHCPNTPEnabled types.Bool `tfsdk:"dhcp_ntp_enabled"`
	DHCPNTP        types.Set  `tfsdk:"dhcp_ntp"`

	// DHCP Boot/PXE
	DHCPBootEnabled  types.Bool   `tfsdk:"dhcp_boot_enabled"`
	DHCPBootServer   types.String `tfsdk:"dhcp_boot_server"`
	DHCPTFTPServer   types.String `tfsdk:"dhcp_tftp_server"`
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
	IPv6 types.Object `tfsdk:"ipv6"`
}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi network/VLAN configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the network.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the network is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the network.",
				Required:    true,
			},
			"purpose": schema.StringAttribute{
				Description: "The purpose of the network. Valid values: 'corporate', 'guest', 'wan', 'vlan-only'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("corporate", "guest", "wan", "vlan-only"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the network is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},

			// VLAN
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID for this network. Must be between 1 and 4095.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 4095),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"subnet": schema.StringAttribute{
				Description: "The subnet in CIDR notation (e.g., '10.0.100.0/24').",
				Optional:    true,
			},

			// DHCP Core
			"dhcp_enabled": schema.BoolAttribute{
				Description: "Whether DHCP is enabled on this network. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"dhcp_start": schema.StringAttribute{
				Description: "The start of the DHCP IP range.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcp_stop": schema.StringAttribute{
				Description: "The end of the DHCP IP range.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcp_lease": schema.Int64Attribute{
				Description: "The DHCP lease time in seconds. Defaults to 86400 (24 hours).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(defaultDHCPLease),
			},

			// DHCP DNS
			"dhcp_dns_enabled": schema.BoolAttribute{
				Description: "Whether to provide DNS servers via DHCP.",
				Optional:    true,
				Computed:    true,
			},
			"dhcp_dns": schema.SetAttribute{
				Description: "Set of DNS servers to provide via DHCP (maximum 4). Must be valid IPv4 addresses.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(4),
					setvalidator.ValueStringsAre(
						IPv4Address(),
					),
				},
			},

			// DHCP Gateway
			"dhcp_gateway_enabled": schema.BoolAttribute{
				Description: "Whether to override the default gateway provided via DHCP (Option 3).",
				Optional:    true,
			},
			"dhcp_gateway": schema.StringAttribute{
				Description: "Custom gateway IP address to provide via DHCP (Option 3). Requires dhcp_gateway_enabled to be true.",
				Optional:    true,
				Validators: []validator.String{
					IPv4Address(),
				},
			},

			// DHCP NTP
			"dhcp_ntp_enabled": schema.BoolAttribute{
				Description: "Whether to provide NTP servers via DHCP (Option 42).",
				Optional:    true,
			},
			"dhcp_ntp": schema.SetAttribute{
				Description: "Set of NTP servers to provide via DHCP (maximum 2). Must be valid IPv4 addresses.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(2),
					setvalidator.ValueStringsAre(
						IPv4Address(),
					),
				},
			},

			// DHCP Boot/PXE
			"dhcp_boot_enabled": schema.BoolAttribute{
				Description: "Whether DHCP network boot (PXE) is enabled. When enabled, DHCP will provide boot options to clients.",
				Optional:    true,
			},
			"dhcp_boot_server": schema.StringAttribute{
				Description: "The IP address of the PXE boot server (DHCP Option 66).",
				Optional:    true,
				Validators: []validator.String{
					IPv4Address(),
				},
			},
			"dhcp_tftp_server": schema.StringAttribute{
				Description: "The TFTP server address (DHCP Option 66 tftp-server-name). If not set, no separate TFTP server is advertised.",
				Optional:    true,
				Validators: []validator.String{
					IPv4Address(),
				},
			},
			"dhcp_boot_filename": schema.StringAttribute{
				Description: "The boot filename to provide to clients (DHCP Option 67). This is the path to the boot file on the TFTP server.",
				Optional:    true,
			},

			// DHCP Additional Options
			"dhcp_relay_enabled": schema.BoolAttribute{
				Description: "Whether DHCP relay is enabled. When enabled, DHCP requests are forwarded to another DHCP server instead of using the built-in server.",
				Optional:    true,
			},
			"dhcp_time_offset_enabled": schema.BoolAttribute{
				Description: "Whether to provide time offset via DHCP (Option 2).",
				Optional:    true,
			},
			"dhcp_unifi_controller": schema.StringAttribute{
				Description: "UniFi controller IP address to provide via DHCP (Option 43). Used for UniFi device adoption.",
				Optional:    true,
				Validators: []validator.String{
					IPv4Address(),
				},
			},
			"dhcp_wpad_url": schema.StringAttribute{
				Description: "Web Proxy Auto-Discovery (WPAD) URL to provide via DHCP (Option 252).",
				Optional:    true,
			},
			"dhcp_guarding_enabled": schema.BoolAttribute{
				Description: "Whether DHCP guarding is enabled. Protects against rogue DHCP servers on the network.",
				Optional:    true,
			},

			// Multicast
			"domain_name": schema.StringAttribute{
				Description: "The domain name for this network.",
				Optional:    true,
			},
			"igmp_snooping": schema.BoolAttribute{
				Description: "Whether IGMP snooping is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"igmp_proxy_upstream": schema.BoolAttribute{
				Description: "Whether this network acts as an IGMP proxy upstream interface.",
				Optional:    true,
			},

			// Network Access
			"internet_access_enabled": schema.BoolAttribute{
				Description: "Whether internet access is enabled for this network. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"intra_network_access_enabled": schema.BoolAttribute{
				Description: "Whether devices on this network can communicate with devices on other networks.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"nat_enabled": schema.BoolAttribute{
				Description: "Whether NAT is enabled for this network. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"mdns_enabled": schema.BoolAttribute{
				Description: "Whether mDNS (Bonjour/Avahi) is enabled for this network.",
				Optional:    true,
			},
			"upnp_lan_enabled": schema.BoolAttribute{
				Description: "Whether UPnP is enabled on this LAN network.",
				Optional:    true,
			},

			// Routing
			"network_group": schema.StringAttribute{
				Description: "The network group. Valid values: 'LAN', 'WAN', 'WAN2'. Defaults to 'LAN'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("LAN"),
				Validators: []validator.String{
					stringvalidator.OneOf("LAN", "WAN", "WAN2"),
				},
			},
			"firewall_zone_id": schema.StringAttribute{
				Description: "The firewall zone ID to associate with this network. If not specified, the controller assigns a default zone.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// IPv6
			"ipv6": schema.SingleNestedAttribute{
				Description: "IPv6 configuration for this network.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"interface_type": schema.StringAttribute{
						Description: "IPv6 interface type. Valid values: 'none', 'static', 'pd'.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("none", "static", "pd"),
						},
					},
					"setting_preference": schema.StringAttribute{
						Description: "IPv6 setting preference. Valid values: 'auto', 'manual'.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("auto", "manual"),
						},
					},
					"wan_delegation_type": schema.StringAttribute{
						Description: "IPv6 WAN delegation type.",
						Optional:    true,
						Computed:    true,
					},
					"subnet": schema.StringAttribute{
						Description: "IPv6 subnet in CIDR notation.",
						Optional:    true,
						Computed:    true,
					},
					"client_address_assignment": schema.StringAttribute{
						Description: "IPv6 client address assignment method.",
						Optional:    true,
						Computed:    true,
					},
					"pd_interface": schema.StringAttribute{
						Description: "IPv6 prefix delegation interface.",
						Optional:    true,
						Computed:    true,
					},
					"pd_prefixid": schema.StringAttribute{
						Description: "IPv6 prefix delegation prefix ID.",
						Optional:    true,
						Computed:    true,
					},
					"pd_start": schema.StringAttribute{
						Description: "IPv6 prefix delegation range start.",
						Optional:    true,
						Computed:    true,
					},
					"pd_stop": schema.StringAttribute{
						Description: "IPv6 prefix delegation range stop.",
						Optional:    true,
						Computed:    true,
					},
					"pd_auto_prefixid_enabled": schema.BoolAttribute{
						Description: "Whether automatic prefix ID assignment is enabled for prefix delegation.",
						Optional:    true,
						Computed:    true,
					},
					"ra_enabled": schema.BoolAttribute{
						Description: "Whether Router Advertisement (RA) is enabled.",
						Optional:    true,
						Computed:    true,
					},
					"ra_preferred_lifetime": schema.Int64Attribute{
						Description: "Router Advertisement preferred lifetime in seconds.",
						Optional:    true,
						Computed:    true,
					},
					"ra_priority": schema.StringAttribute{
						Description: "Router Advertisement priority. Valid values: 'high', 'medium', 'low'.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("high", "medium", "low"),
						},
					},
					"ra_valid_lifetime": schema.Int64Attribute{
						Description: "Router Advertisement valid lifetime in seconds.",
						Optional:    true,
						Computed:    true,
					},
					"dhcpv6_enabled": schema.BoolAttribute{
						Description: "Whether DHCPv6 is enabled.",
						Optional:    true,
						Computed:    true,
					},
					"dhcpv6_start": schema.StringAttribute{
						Description: "DHCPv6 range start address.",
						Optional:    true,
						Computed:    true,
					},
					"dhcpv6_stop": schema.StringAttribute{
						Description: "DHCPv6 range stop address.",
						Optional:    true,
						Computed:    true,
					},
					"dhcpv6_lease_time": schema.Int64Attribute{
						Description: "DHCPv6 lease time in seconds.",
						Optional:    true,
						Computed:    true,
					},
					"dhcpv6_dns_auto": schema.BoolAttribute{
						Description: "Whether DHCPv6 DNS is set automatically.",
						Optional:    true,
						Computed:    true,
					},
					"dhcpv6_dns": schema.ListAttribute{
						Description: "DHCPv6 DNS servers (maximum 4).",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtMost(4),
						},
					},
					"dhcpv6_allow_slaac": schema.BoolAttribute{
						Description: "Whether to allow SLAAC alongside DHCPv6.",
						Optional:    true,
						Computed:    true,
					},
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

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkResourceModel

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

	network := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateNetwork(ctx, network)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "network")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkResourceModel

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

	network, err := r.client.GetNetwork(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "network")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, network, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NetworkResourceModel
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

	network := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	network.ID = state.ID.ValueString()
	network.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateNetwork(ctx, state.ID.ValueString(), network)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "network")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkResourceModel

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

	err := r.client.DeleteNetwork(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "network")
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkResource) planToSDK(ctx context.Context, plan *NetworkResourceModel, diags *diag.Diagnostics) *unifi.Network {
	network := &unifi.Network{
		Name:    plan.Name.ValueString(),
		Purpose: plan.Purpose.ValueString(),
		Enabled: boolPtr(plan.Enabled.ValueBool()),
	}

	// VLAN
	if !plan.VlanID.IsNull() && !plan.VlanID.IsUnknown() {
		network.VLAN = intPtr(plan.VlanID.ValueInt64())
		network.VLANEnabled = boolPtr(true)
	}
	if !plan.Subnet.IsNull() {
		network.IPSubnet = plan.Subnet.ValueString()
	}

	// DHCP Core
	network.DHCPDEnabled = boolPtr(plan.DHCPEnabled.ValueBool())
	if !plan.DHCPStart.IsNull() {
		network.DHCPDStart = plan.DHCPStart.ValueString()
	}
	if !plan.DHCPStop.IsNull() {
		network.DHCPDStop = plan.DHCPStop.ValueString()
	}
	if !plan.DHCPLease.IsNull() && !plan.DHCPLease.IsUnknown() {
		network.DHCPDLeasetime = intPtr(plan.DHCPLease.ValueInt64())
	}

	// DHCP DNS
	if !plan.DHCPDNSEnabled.IsNull() {
		network.DHCPDDNSEnabled = boolPtr(plan.DHCPDNSEnabled.ValueBool())
	}
	if !plan.DHCPDNS.IsNull() {
		var dnsServers []string
		diags.Append(plan.DHCPDNS.ElementsAs(ctx, &dnsServers, false)...)
		if diags.HasError() {
			return nil
		}
		if len(dnsServers) > 0 {
			network.DHCPDDns1 = dnsServers[0]
		}
		if len(dnsServers) > 1 {
			network.DHCPDDns2 = dnsServers[1]
		}
		if len(dnsServers) > 2 {
			network.DHCPDDns3 = dnsServers[2]
		}
		if len(dnsServers) > 3 {
			network.DHCPDDns4 = dnsServers[3]
		}
	}

	// DHCP Gateway
	if !plan.DHCPGatewayEnabled.IsNull() {
		network.DHCPDGatewayEnabled = boolPtr(plan.DHCPGatewayEnabled.ValueBool())
	}
	if !plan.DHCPGateway.IsNull() {
		network.DHCPDGateway = plan.DHCPGateway.ValueString()
	}

	// DHCP NTP
	if !plan.DHCPNTPEnabled.IsNull() {
		network.DHCPDNTPEnabled = boolPtr(plan.DHCPNTPEnabled.ValueBool())
	}
	if !plan.DHCPNTP.IsNull() {
		var ntpServers []string
		diags.Append(plan.DHCPNTP.ElementsAs(ctx, &ntpServers, false)...)
		if diags.HasError() {
			return nil
		}
		if len(ntpServers) > 0 {
			network.DHCPDNtp1 = ntpServers[0]
		}
		if len(ntpServers) > 1 {
			network.DHCPDNtp2 = ntpServers[1]
		}
	}

	// DHCP Boot/PXE
	if !plan.DHCPBootEnabled.IsNull() {
		network.DHCPDBootEnabled = boolPtr(plan.DHCPBootEnabled.ValueBool())
	}
	if !plan.DHCPBootServer.IsNull() {
		network.DHCPDBootServer = plan.DHCPBootServer.ValueString()
	}
	if !plan.DHCPTFTPServer.IsNull() {
		network.DHCPDTFTPServer = plan.DHCPTFTPServer.ValueString()
	}
	if !plan.DHCPBootFilename.IsNull() {
		network.DHCPDBootFilename = plan.DHCPBootFilename.ValueString()
	}

	// DHCP Additional Options
	if !plan.DHCPRelayEnabled.IsNull() {
		network.DHCPRelayEnabled = boolPtr(plan.DHCPRelayEnabled.ValueBool())
	}
	if !plan.DHCPTimeOffsetEnabled.IsNull() {
		network.DHCPDTimeOffsetEnabled = boolPtr(plan.DHCPTimeOffsetEnabled.ValueBool())
	}
	if !plan.DHCPUnifiController.IsNull() {
		network.DHCPDUnifiController = plan.DHCPUnifiController.ValueString()
	}
	if !plan.DHCPWPADUrl.IsNull() {
		network.DHCPDWPADUrl = plan.DHCPWPADUrl.ValueString()
	}
	if !plan.DHCPGuardingEnabled.IsNull() {
		network.DHCPGuardingEnabled = boolPtr(plan.DHCPGuardingEnabled.ValueBool())
	}

	// Multicast
	if !plan.DomainName.IsNull() {
		network.DomainName = plan.DomainName.ValueString()
	}
	network.IGMPSnooping = boolPtr(plan.IGMPSnooping.ValueBool())
	if !plan.IGMPProxyUpstream.IsNull() {
		network.IGMPProxyUpstream = boolPtr(plan.IGMPProxyUpstream.ValueBool())
	}

	// Network Access
	network.InternetAccessEnabled = boolPtr(plan.InternetAccessEnabled.ValueBool())
	network.IntraNetworkAccessEnabled = boolPtr(plan.IntraNetworkAccessEnabled.ValueBool())
	network.IsNAT = boolPtr(plan.NATEnabled.ValueBool())
	if !plan.MDNSEnabled.IsNull() {
		network.MDNSEnabled = boolPtr(plan.MDNSEnabled.ValueBool())
	}
	if !plan.UPnPLANEnabled.IsNull() {
		network.UpnpLANEnabled = boolPtr(plan.UPnPLANEnabled.ValueBool())
	}

	// Routing
	network.NetworkGroup = plan.NetworkGroup.ValueString()
	if !plan.FirewallZoneID.IsNull() {
		network.FirewallZoneID = plan.FirewallZoneID.ValueString()
	}

	// IPv6
	if !plan.IPv6.IsNull() && !plan.IPv6.IsUnknown() {
		ipv6FromObject(ctx, plan.IPv6, network, diags)
		if diags.HasError() {
			return nil
		}
	}

	return network
}

func (r *NetworkResource) sdkToState(ctx context.Context, network *unifi.Network, state *NetworkResourceModel) diag.Diagnostics {
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
		state.DHCPLease = types.Int64Value(defaultDHCPLease)
	}

	// DHCP DNS
	if network.DHCPDDNSEnabled != nil {
		state.DHCPDNSEnabled = types.BoolValue(*network.DHCPDDNSEnabled)
	} else {
		state.DHCPDNSEnabled = types.BoolNull()
	}
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
	state.DHCPTFTPServer = stringValueOrNull(network.DHCPDTFTPServer)
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
	ipv6Obj, d := ipv6ToObject(ctx, network)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	state.IPv6 = ipv6Obj

	return diags
}

func ipv6FromObject(ctx context.Context, obj types.Object, network *unifi.Network, diags *diag.Diagnostics) {
	attrs := obj.Attributes()

	if v, ok := attrs["interface_type"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6InterfaceType = v.ValueString()
	}
	if v, ok := attrs["setting_preference"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPv6SettingPreference = v.ValueString()
	}
	if v, ok := attrs["wan_delegation_type"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPv6WANDelegationType = v.ValueString()
	}
	if v, ok := attrs["subnet"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6Subnet = v.ValueString()
	}
	if v, ok := attrs["client_address_assignment"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6ClientAddressAssignment = v.ValueString()
	}
	if v, ok := attrs["pd_interface"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6PDInterface = v.ValueString()
	}
	if v, ok := attrs["pd_prefixid"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6PDPrefixid = v.ValueString()
	}
	if v, ok := attrs["pd_start"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6PDStart = v.ValueString()
	}
	if v, ok := attrs["pd_stop"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6PDStop = v.ValueString()
	}
	if v, ok := attrs["pd_auto_prefixid_enabled"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6PDAutoPrefixidEnabled = boolPtr(v.ValueBool())
	}
	if v, ok := attrs["ra_enabled"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6RaEnabled = boolPtr(v.ValueBool())
	}
	if v, ok := attrs["ra_preferred_lifetime"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6RaPreferredLifetime = intPtr(v.ValueInt64())
	}
	if v, ok := attrs["ra_priority"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6RaPriority = v.ValueString()
	}
	if v, ok := attrs["ra_valid_lifetime"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		network.IPV6RaValidLifetime = intPtr(v.ValueInt64())
	}
	if v, ok := attrs["dhcpv6_enabled"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		network.DHCPDV6Enabled = boolPtr(v.ValueBool())
	}
	if v, ok := attrs["dhcpv6_start"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.DHCPDV6Start = v.ValueString()
	}
	if v, ok := attrs["dhcpv6_stop"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		network.DHCPDV6Stop = v.ValueString()
	}
	if v, ok := attrs["dhcpv6_lease_time"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		network.DHCPDV6LeaseTime = intPtr(v.ValueInt64())
	}
	if v, ok := attrs["dhcpv6_dns_auto"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		network.DHCPDV6DNSAuto = boolPtr(v.ValueBool())
	}
	if v, ok := attrs["dhcpv6_dns"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		var dnsServers []string
		diags.Append(v.ElementsAs(ctx, &dnsServers, false)...)
		if diags.HasError() {
			return
		}
		if len(dnsServers) > 0 {
			network.DHCPDV6DNS1 = dnsServers[0]
		}
		if len(dnsServers) > 1 {
			network.DHCPDV6DNS2 = dnsServers[1]
		}
		if len(dnsServers) > 2 {
			network.DHCPDV6DNS3 = dnsServers[2]
		}
		if len(dnsServers) > 3 {
			network.DHCPDV6DNS4 = dnsServers[3]
		}
	}
	if v, ok := attrs["dhcpv6_allow_slaac"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		network.DHCPDV6AllowSlaac = boolPtr(v.ValueBool())
	}
}

func ipv6ToObject(ctx context.Context, network *unifi.Network) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Collect DHCPv6 DNS servers into a list
	var dnsVal types.List
	var dnsServers []string
	if network.DHCPDV6DNS1 != "" {
		dnsServers = append(dnsServers, network.DHCPDV6DNS1)
	}
	if network.DHCPDV6DNS2 != "" {
		dnsServers = append(dnsServers, network.DHCPDV6DNS2)
	}
	if network.DHCPDV6DNS3 != "" {
		dnsServers = append(dnsServers, network.DHCPDV6DNS3)
	}
	if network.DHCPDV6DNS4 != "" {
		dnsServers = append(dnsServers, network.DHCPDV6DNS4)
	}
	if len(dnsServers) > 0 {
		dnsList, d := types.ListValueFrom(ctx, types.StringType, dnsServers)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(ipv6AttrTypes), diags
		}
		dnsVal = dnsList
	} else {
		dnsVal = types.ListNull(types.StringType)
	}

	attrs := map[string]attr.Value{
		"interface_type":             stringValueOrNull(network.IPV6InterfaceType),
		"setting_preference":         stringValueOrNull(network.IPv6SettingPreference),
		"wan_delegation_type":        stringValueOrNull(network.IPv6WANDelegationType),
		"subnet":                     stringValueOrNull(network.IPV6Subnet),
		"client_address_assignment":  stringValueOrNull(network.IPV6ClientAddressAssignment),
		"pd_interface":               stringValueOrNull(network.IPV6PDInterface),
		"pd_prefixid":                stringValueOrNull(network.IPV6PDPrefixid),
		"pd_start":                   stringValueOrNull(network.IPV6PDStart),
		"pd_stop":                    stringValueOrNull(network.IPV6PDStop),
		"pd_auto_prefixid_enabled":   types.BoolValue(derefBool(network.IPV6PDAutoPrefixidEnabled)),
		"ra_enabled":                 types.BoolValue(derefBool(network.IPV6RaEnabled)),
		"ra_preferred_lifetime":      types.Int64Value(derefInt(network.IPV6RaPreferredLifetime)),
		"ra_priority":                stringValueOrNull(network.IPV6RaPriority),
		"ra_valid_lifetime":          types.Int64Value(derefInt(network.IPV6RaValidLifetime)),
		"dhcpv6_enabled":             types.BoolValue(derefBool(network.DHCPDV6Enabled)),
		"dhcpv6_start":               stringValueOrNull(network.DHCPDV6Start),
		"dhcpv6_stop":                stringValueOrNull(network.DHCPDV6Stop),
		"dhcpv6_lease_time":          types.Int64Value(derefInt(network.DHCPDV6LeaseTime)),
		"dhcpv6_dns_auto":            types.BoolValue(derefBool(network.DHCPDV6DNSAuto)),
		"dhcpv6_dns":                 dnsVal,
		"dhcpv6_allow_slaac":         types.BoolValue(derefBool(network.DHCPDV6AllowSlaac)),
	}

	obj, d := types.ObjectValue(ipv6AttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}
