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

var _ datasource.DataSource = &StaticRouteDataSource{}

type StaticRouteDataSource struct {
	client *AutoLoginClient
}

type StaticRouteDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	SiteID               types.String `tfsdk:"site_id"`
	Name                 types.String `tfsdk:"name"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Type                 types.String `tfsdk:"type"`
	StaticRouteNetwork   types.String `tfsdk:"static_route_network"`
	StaticRouteNexthop   types.String `tfsdk:"static_route_nexthop"`
	StaticRouteDistance  types.Int64  `tfsdk:"static_route_distance"`
	StaticRouteInterface types.String `tfsdk:"static_route_interface"`
	StaticRouteType      types.String `tfsdk:"static_route_type"`
}

func NewStaticRouteDataSource() datasource.DataSource {
	return &StaticRouteDataSource{}
}

func (d *StaticRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_route"
}

func (d *StaticRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi static route. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the static route. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the static route. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the static route exists.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the static route is enabled.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of route (static-route, interface-route).",
				Computed:    true,
			},
			"static_route_network": schema.StringAttribute{
				Description: "The destination network in CIDR notation.",
				Computed:    true,
			},
			"static_route_nexthop": schema.StringAttribute{
				Description: "The next hop IP address for the route.",
				Computed:    true,
			},
			"static_route_distance": schema.Int64Attribute{
				Description: "The administrative distance (metric) for the route.",
				Computed:    true,
			},
			"static_route_interface": schema.StringAttribute{
				Description: "The interface for the route (for interface routes).",
				Computed:    true,
			},
			"static_route_type": schema.StringAttribute{
				Description: "The static route type (nexthop-route, interface-route, blackhole).",
				Computed:    true,
			},
		},
	}
}

func (d *StaticRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StaticRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StaticRouteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a static route.",
		)
		return
	}

	var route *unifi.Routing
	var err error

	if hasID {
		route, err = d.client.GetRoute(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "static route")
			return
		}
	} else {
		routes, err := d.client.ListRoutes(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "static routes")
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
				"Static Route Not Found",
				fmt.Sprintf("No static route found with name '%s'.", searchName),
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

func (d *StaticRouteDataSource) sdkToState(ctx context.Context, route *unifi.Routing, state *StaticRouteDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(route.ID)
	state.SiteID = types.StringValue(route.SiteID)
	state.Name = types.StringValue(route.Name)
	state.Enabled = types.BoolValue(derefBool(route.Enabled))
	state.Type = types.StringValue(route.Type)
	state.StaticRouteNetwork = types.StringValue(route.StaticRouteNetwork)
	state.StaticRouteType = types.StringValue(route.StaticRouteType)

	if route.StaticRouteNexthop != "" {
		state.StaticRouteNexthop = types.StringValue(route.StaticRouteNexthop)
	} else {
		state.StaticRouteNexthop = types.StringNull()
	}

	if route.StaticRouteDistance != nil {
		state.StaticRouteDistance = types.Int64Value(int64(*route.StaticRouteDistance))
	} else {
		state.StaticRouteDistance = types.Int64Null()
	}

	if route.StaticRouteInterface != "" {
		state.StaticRouteInterface = types.StringValue(route.StaticRouteInterface)
	} else {
		state.StaticRouteInterface = types.StringNull()
	}

	return diags
}
