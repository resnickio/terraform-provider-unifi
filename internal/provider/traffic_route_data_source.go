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

var _ datasource.DataSource = &TrafficRouteDataSource{}

type TrafficRouteDataSource struct {
	client *AutoLoginClient
}

type TrafficRouteDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Description    types.String `tfsdk:"description"`
	MatchingTarget types.String `tfsdk:"matching_target"`
	TargetDevices  types.List   `tfsdk:"target_devices"`
	NetworkID      types.String `tfsdk:"network_id"`
	Domains        types.List   `tfsdk:"domains"`
	IPAddresses    types.Set    `tfsdk:"ip_addresses"`
	IPRanges       types.Set    `tfsdk:"ip_ranges"`
	Regions        types.Set    `tfsdk:"regions"`
	Fallback       types.Bool   `tfsdk:"fallback"`
	KillSwitch     types.Bool   `tfsdk:"kill_switch"`
}

func NewTrafficRouteDataSource() datasource.DataSource {
	return &TrafficRouteDataSource{}
}

func (d *TrafficRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_route"
}

func (d *TrafficRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi traffic route (policy-based routing). Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the traffic route. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the traffic route. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the traffic route is enabled.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the traffic route.",
				Computed:    true,
			},
			"matching_target": schema.StringAttribute{
				Description: "The matching target type (INTERNET, IP, DOMAIN, REGION, APP).",
				Computed:    true,
			},
			"target_devices": schema.ListNestedAttribute{
				Description: "List of target devices for the route.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_mac": schema.StringAttribute{
							Description: "The MAC address of the client device.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The target type (ALL_CLIENTS, CLIENT, NETWORK).",
							Computed:    true,
						},
						"network_id": schema.StringAttribute{
							Description: "The network ID for network-based targeting.",
							Computed:    true,
						},
					},
				},
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID to route traffic through.",
				Computed:    true,
			},
			"domains": schema.ListNestedAttribute{
				Description: "List of domains for domain-based routing.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The domain name or pattern.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description for the domain entry.",
							Computed:    true,
						},
						"ports": schema.SetAttribute{
							Description: "Set of ports associated with the domain.",
							Computed:    true,
							ElementType: types.Int64Type,
						},
					},
				},
			},
			"ip_addresses": schema.SetAttribute{
				Description: "Set of IP addresses or CIDR blocks for IP-based routing.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"ip_ranges": schema.SetAttribute{
				Description: "Set of IP ranges for IP-based routing.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"regions": schema.SetAttribute{
				Description: "Set of geographic regions for region-based routing.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"fallback": schema.BoolAttribute{
				Description: "Whether to use fallback routing.",
				Computed:    true,
			},
			"kill_switch": schema.BoolAttribute{
				Description: "Whether kill switch is enabled (block traffic if VPN fails).",
				Computed:    true,
			},
		},
	}
}

func (d *TrafficRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TrafficRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config TrafficRouteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a traffic route.",
		)
		return
	}

	var route *unifi.TrafficRoute
	var err error

	if hasID {
		route, err = d.client.GetTrafficRoute(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "traffic route")
			return
		}
	} else {
		routes, err := d.client.ListTrafficRoutes(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "traffic routes")
			return
		}

		searchName := config.Name.ValueString()
		for i := range routes {
			if routes[i].Name == searchName {
				route = &routes[i]
				break
			}
		}

		if route == nil {
			resp.Diagnostics.AddError(
				"Traffic Route Not Found",
				fmt.Sprintf("No traffic route found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, route, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *TrafficRouteDataSource) sdkToState(ctx context.Context, route *unifi.TrafficRoute, state *TrafficRouteDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(route.ID)
	state.Name = types.StringValue(route.Name)
	state.Enabled = types.BoolValue(derefBool(route.Enabled))
	state.Fallback = types.BoolValue(derefBool(route.Fallback))
	state.KillSwitch = types.BoolValue(derefBool(route.KillSwitch))

	if route.Description != "" {
		state.Description = types.StringValue(route.Description)
	} else {
		state.Description = types.StringNull()
	}

	if route.MatchingTarget != "" {
		state.MatchingTarget = types.StringValue(route.MatchingTarget)
	} else {
		state.MatchingTarget = types.StringNull()
	}

	if route.NetworkID != "" {
		state.NetworkID = types.StringValue(route.NetworkID)
	} else {
		state.NetworkID = types.StringNull()
	}

	targets, diagsTargets := trafficTargetsToList(ctx, route.TargetDevices)
	diags.Append(diagsTargets...)
	state.TargetDevices = targets

	domains, diagsDomains := trafficDomainsToList(ctx, route.Domains)
	diags.Append(diagsDomains...)
	state.Domains = domains

	if len(route.IPAddresses) > 0 {
		ips, d := types.SetValueFrom(ctx, types.StringType, route.IPAddresses)
		diags.Append(d...)
		state.IPAddresses = ips
	} else {
		state.IPAddresses = types.SetNull(types.StringType)
	}

	if len(route.IPRanges) > 0 {
		ranges, d := types.SetValueFrom(ctx, types.StringType, route.IPRanges)
		diags.Append(d...)
		state.IPRanges = ranges
	} else {
		state.IPRanges = types.SetNull(types.StringType)
	}

	if len(route.Regions) > 0 {
		regions, d := types.SetValueFrom(ctx, types.StringType, route.Regions)
		diags.Append(d...)
		state.Regions = regions
	} else {
		state.Regions = types.SetNull(types.StringType)
	}

	return diags
}
