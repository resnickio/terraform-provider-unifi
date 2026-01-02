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
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	SiteID       types.String `tfsdk:"site_id"`
	Purpose      types.String `tfsdk:"purpose"`
	VlanID       types.Int64  `tfsdk:"vlan_id"`
	NetworkGroup types.String `tfsdk:"network_group"`
	Subnet       types.String `tfsdk:"subnet"`
	DHCPEnabled  types.Bool   `tfsdk:"dhcp_enabled"`
	DHCPStart    types.String `tfsdk:"dhcp_start"`
	DHCPStop     types.String `tfsdk:"dhcp_stop"`
	DHCPLease    types.Int64  `tfsdk:"dhcp_lease"`
	DHCPDNS      types.Set    `tfsdk:"dhcp_dns"`
	DomainName   types.String `tfsdk:"domain_name"`
	IGMPSnooping types.Bool   `tfsdk:"igmp_snooping"`
	Enabled      types.Bool   `tfsdk:"enabled"`
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
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID for this network.",
				Computed:    true,
			},
			"network_group": schema.StringAttribute{
				Description: "The network group.",
				Computed:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "The subnet in CIDR notation.",
				Computed:    true,
			},
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
			"dhcp_dns": schema.SetAttribute{
				Description: "Set of DNS servers provided via DHCP.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name for this network.",
				Computed:    true,
			},
			"igmp_snooping": schema.BoolAttribute{
				Description: "Whether IGMP snooping is enabled.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the network is enabled.",
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
	state.NetworkGroup = types.StringValue(network.NetworkGroup)

	if network.VLAN != nil {
		state.VlanID = types.Int64Value(int64(*network.VLAN))
	} else {
		state.VlanID = types.Int64Null()
	}

	if network.IPSubnet != "" {
		state.Subnet = types.StringValue(network.IPSubnet)
	} else {
		state.Subnet = types.StringNull()
	}

	state.DHCPEnabled = types.BoolValue(derefBool(network.DHCPDEnabled))

	if network.DHCPDStart != "" {
		state.DHCPStart = types.StringValue(network.DHCPDStart)
	} else {
		state.DHCPStart = types.StringNull()
	}

	if network.DHCPDStop != "" {
		state.DHCPStop = types.StringValue(network.DHCPDStop)
	} else {
		state.DHCPStop = types.StringNull()
	}

	if network.DHCPDLeasetime != nil {
		state.DHCPLease = types.Int64Value(int64(*network.DHCPDLeasetime))
	} else {
		state.DHCPLease = types.Int64Null()
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

	if network.DomainName != "" {
		state.DomainName = types.StringValue(network.DomainName)
	} else {
		state.DomainName = types.StringNull()
	}

	state.IGMPSnooping = types.BoolValue(derefBool(network.IGMPSnooping))
	state.Enabled = types.BoolValue(derefBool(network.Enabled))

	return diags
}
