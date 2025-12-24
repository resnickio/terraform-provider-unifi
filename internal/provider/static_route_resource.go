package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &StaticRouteResource{}
	_ resource.ResourceWithImportState = &StaticRouteResource{}
)

type StaticRouteResource struct {
	client *AutoLoginClient
}

type StaticRouteResourceModel struct {
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

func NewStaticRouteResource() resource.Resource {
	return &StaticRouteResource{}
}

func (r *StaticRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_route"
}

func (r *StaticRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi static route.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the static route.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the static route is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the static route.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the static route is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The type of route. Valid values: 'static-route', 'interface-route'. Defaults to 'static-route'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("static-route"),
			},
			"static_route_network": schema.StringAttribute{
				Description: "The destination network in CIDR notation (e.g., '10.0.0.0/24').",
				Required:    true,
			},
			"static_route_nexthop": schema.StringAttribute{
				Description: "The next hop IP address for the route.",
				Optional:    true,
			},
			"static_route_distance": schema.Int64Attribute{
				Description: "The administrative distance (metric) for the route.",
				Optional:    true,
			},
			"static_route_interface": schema.StringAttribute{
				Description: "The interface for the route (for interface routes).",
				Optional:    true,
			},
			"static_route_type": schema.StringAttribute{
				Description: "The static route type. Valid values: 'nexthop-route', 'interface-route', 'blackhole'. Defaults to 'nexthop-route'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("nexthop-route"),
			},
		},
	}
}

func (r *StaticRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StaticRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route := r.planToSDK(&plan)

	created, err := r.client.CreateRoute(ctx, route)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "static route")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StaticRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.GetRoute(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "static route")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(route, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StaticRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state StaticRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route := r.planToSDK(&plan)
	route.ID = state.ID.ValueString()
	route.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateRoute(ctx, state.ID.ValueString(), route)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "static route")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StaticRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRoute(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "static route")
		return
	}
}

func (r *StaticRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *StaticRouteResource) planToSDK(plan *StaticRouteResourceModel) *unifi.Routing {
	route := &unifi.Routing{
		Name:               plan.Name.ValueString(),
		Enabled:            boolPtr(plan.Enabled.ValueBool()),
		Type:               plan.Type.ValueString(),
		StaticRouteNetwork: plan.StaticRouteNetwork.ValueString(),
		StaticRouteType:    plan.StaticRouteType.ValueString(),
	}

	if !plan.StaticRouteNexthop.IsNull() && !plan.StaticRouteNexthop.IsUnknown() {
		route.StaticRouteNexthop = plan.StaticRouteNexthop.ValueString()
	}

	if !plan.StaticRouteDistance.IsNull() && !plan.StaticRouteDistance.IsUnknown() {
		route.StaticRouteDistance = intPtr(plan.StaticRouteDistance.ValueInt64())
	}

	if !plan.StaticRouteInterface.IsNull() && !plan.StaticRouteInterface.IsUnknown() {
		route.StaticRouteInterface = plan.StaticRouteInterface.ValueString()
	}

	return route
}

func (r *StaticRouteResource) sdkToState(route *unifi.Routing, state *StaticRouteResourceModel) diag.Diagnostics {
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
