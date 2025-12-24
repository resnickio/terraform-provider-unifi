package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type NetworkResource struct {
	client *AutoLoginClient
}

type NetworkResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SiteID       types.String `tfsdk:"site_id"`
	Name         types.String `tfsdk:"name"`
	Purpose      types.String `tfsdk:"purpose"`
	VlanID       types.Int64  `tfsdk:"vlan_id"`
	NetworkGroup types.String `tfsdk:"network_group"`
	Subnet       types.String `tfsdk:"subnet"`
	DHCPEnabled  types.Bool   `tfsdk:"dhcp_enabled"`
	DHCPStart    types.String `tfsdk:"dhcp_start"`
	DHCPStop     types.String `tfsdk:"dhcp_stop"`
	DHCPLease    types.Int64  `tfsdk:"dhcp_lease"`
	DHCPDNS      types.List   `tfsdk:"dhcp_dns"`
	DomainName   types.String `tfsdk:"domain_name"`
	IGMPSnooping types.Bool   `tfsdk:"igmp_snooping"`
	Enabled      types.Bool   `tfsdk:"enabled"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID for this network. Must be between 1 and 4095.",
				Optional:    true,
				Computed:    true,
			},
			"network_group": schema.StringAttribute{
				Description: "The network group. Defaults to 'LAN'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("LAN"),
			},
			"subnet": schema.StringAttribute{
				Description: "The subnet in CIDR notation (e.g., '10.0.100.0/24').",
				Optional:    true,
			},
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
			},
			"dhcp_stop": schema.StringAttribute{
				Description: "The end of the DHCP IP range.",
				Optional:    true,
				Computed:    true,
			},
			"dhcp_lease": schema.Int64Attribute{
				Description: "The DHCP lease time in seconds. Defaults to 86400 (24 hours).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(defaultDHCPLease),
			},
			"dhcp_dns": schema.ListAttribute{
				Description: "List of DNS servers to provide via DHCP.",
				Optional:    true,
				ElementType: types.StringType,
			},
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
			"enabled": schema.BoolAttribute{
				Description: "Whether the network is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
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

	// Convert plan to SDK struct
	network := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the network
	created, err := r.client.CreateNetwork(ctx, network)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "network")
		return
	}

	// Update state with response
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

	// Get the network
	network, err := r.client.GetNetwork(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "network")
		return
	}

	// Update state with response
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

	// Convert plan to SDK struct
	network := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID and SiteID from state
	network.ID = state.ID.ValueString()
	network.SiteID = state.SiteID.ValueString()

	// Update the network
	updated, err := r.client.UpdateNetwork(ctx, state.ID.ValueString(), network)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "network")
		return
	}

	// Update state with response
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

	// Delete the network
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

// planToSDK converts the Terraform plan to an SDK Network struct.
func (r *NetworkResource) planToSDK(ctx context.Context, plan *NetworkResourceModel, diags *diag.Diagnostics) *unifi.Network {
	network := &unifi.Network{
		Name:    plan.Name.ValueString(),
		Purpose: plan.Purpose.ValueString(),
		Enabled: boolPtr(plan.Enabled.ValueBool()),
	}

	// NetworkRouting fields (embedded)
	network.NetworkGroup = plan.NetworkGroup.ValueString()

	// NetworkDHCP fields (embedded)
	network.DHCPDEnabled = boolPtr(plan.DHCPEnabled.ValueBool())

	// NetworkMulticast fields (embedded)
	network.IGMPSnooping = boolPtr(plan.IGMPSnooping.ValueBool())

	// NetworkVLAN fields (embedded)
	if !plan.VlanID.IsNull() && !plan.VlanID.IsUnknown() {
		network.VLAN = intPtr(plan.VlanID.ValueInt64())
		network.VLANEnabled = boolPtr(true)
	}

	if !plan.Subnet.IsNull() {
		network.IPSubnet = plan.Subnet.ValueString()
	}

	if !plan.DHCPStart.IsNull() {
		network.DHCPDStart = plan.DHCPStart.ValueString()
	}

	if !plan.DHCPStop.IsNull() {
		network.DHCPDStop = plan.DHCPStop.ValueString()
	}

	if !plan.DHCPLease.IsNull() && !plan.DHCPLease.IsUnknown() {
		network.DHCPDLeasetime = intPtr(plan.DHCPLease.ValueInt64())
	}

	if !plan.DHCPDNS.IsNull() {
		var dnsServers []string
		diags.Append(plan.DHCPDNS.ElementsAs(ctx, &dnsServers, false)...)
		if diags.HasError() {
			return nil
		}
		network.DHCPDDNSEnabled = boolPtr(len(dnsServers) > 0)
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

	if !plan.DomainName.IsNull() {
		network.DomainName = plan.DomainName.ValueString()
	}

	return network
}

// sdkToState updates the Terraform state from an SDK Network struct.
func (r *NetworkResource) sdkToState(ctx context.Context, network *unifi.Network, state *NetworkResourceModel) diag.Diagnostics {
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
		state.DHCPLease = types.Int64Value(defaultDHCPLease)
	}

	// Collect DNS servers
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
		dnsList, d := types.ListValueFrom(ctx, types.StringType, dnsServers)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.DHCPDNS = dnsList
	} else {
		state.DHCPDNS = types.ListNull(types.StringType)
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
