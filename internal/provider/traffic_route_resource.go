package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &TrafficRouteResource{}
	_ resource.ResourceWithImportState = &TrafficRouteResource{}
)

type TrafficRouteResource struct {
	client *AutoLoginClient
}

type TrafficRouteResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	Enabled        types.Bool     `tfsdk:"enabled"`
	Description    types.String   `tfsdk:"description"`
	MatchingTarget types.String   `tfsdk:"matching_target"`
	TargetDevices  types.List     `tfsdk:"target_devices"`
	NetworkID      types.String   `tfsdk:"network_id"`
	Domains        types.List     `tfsdk:"domains"`
	IPAddresses    types.Set      `tfsdk:"ip_addresses"`
	IPRanges       types.Set      `tfsdk:"ip_ranges"`
	Regions        types.Set      `tfsdk:"regions"`
	Fallback       types.Bool     `tfsdk:"fallback"`
	KillSwitch     types.Bool     `tfsdk:"kill_switch"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func NewTrafficRouteResource() resource.Resource {
	return &TrafficRouteResource{}
}

func (r *TrafficRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_route"
}

func (r *TrafficRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi traffic route for policy-based routing.",
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the traffic route.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the traffic route.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the traffic route is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				Description: "A description for the traffic route.",
				Optional:    true,
			},
			"matching_target": schema.StringAttribute{
				Description: "The matching target type. Valid values: INTERNET, IP, DOMAIN, REGION, APP.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("INTERNET", "IP", "DOMAIN", "REGION", "APP"),
				},
			},
			"target_devices": schema.ListNestedAttribute{
				Description: "List of target devices for the route.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_mac": schema.StringAttribute{
							Description: "The MAC address of the client device.",
							Optional:    true,
						},
						"type": schema.StringAttribute{
							Description: "The target type (e.g., ALL_CLIENTS, CLIENT, NETWORK).",
							Optional:    true,
						},
						"network_id": schema.StringAttribute{
							Description: "The network ID for network-based targeting.",
							Optional:    true,
						},
					},
				},
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID to route traffic through.",
				Optional:    true,
			},
			"domains": schema.ListNestedAttribute{
				Description: "List of domains for domain-based routing.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The domain name or pattern.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description for the domain entry.",
							Optional:    true,
						},
						"ports": schema.SetAttribute{
							Description: "Set of ports associated with the domain.",
							Optional:    true,
							ElementType: types.Int64Type,
						},
					},
				},
			},
			"ip_addresses": schema.SetAttribute{
				Description: "Set of IP addresses or CIDR blocks for IP-based routing.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ip_ranges": schema.SetAttribute{
				Description: "Set of IP ranges for IP-based routing.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"regions": schema.SetAttribute{
				Description: "Set of geographic regions for region-based routing.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"fallback": schema.BoolAttribute{
				Description: "Whether to use fallback routing. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"kill_switch": schema.BoolAttribute{
				Description: "Whether to enable kill switch (block traffic if VPN fails). Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *TrafficRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TrafficRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TrafficRouteResourceModel

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

	route := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTrafficRoute(ctx, route)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "traffic route")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrafficRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TrafficRouteResourceModel

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

	route, err := r.client.GetTrafficRoute(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "traffic route")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, route, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TrafficRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TrafficRouteResourceModel

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

	var state TrafficRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	route.ID = state.ID.ValueString()

	updated, err := r.client.UpdateTrafficRoute(ctx, state.ID.ValueString(), route)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "traffic route")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrafficRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TrafficRouteResourceModel

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

	err := r.client.DeleteTrafficRoute(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "traffic route")
		return
	}
}

func (r *TrafficRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TrafficRouteResource) planToSDK(ctx context.Context, plan *TrafficRouteResourceModel, diags *diag.Diagnostics) *unifi.TrafficRoute {
	route := &unifi.TrafficRoute{
		Name:       plan.Name.ValueString(),
		Enabled:    boolPtr(plan.Enabled.ValueBool()),
		Fallback:   boolPtr(plan.Fallback.ValueBool()),
		KillSwitch: boolPtr(plan.KillSwitch.ValueBool()),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		route.Description = plan.Description.ValueString()
	}

	if !plan.MatchingTarget.IsNull() && !plan.MatchingTarget.IsUnknown() {
		route.MatchingTarget = plan.MatchingTarget.ValueString()
	}

	if !plan.NetworkID.IsNull() && !plan.NetworkID.IsUnknown() {
		route.NetworkID = plan.NetworkID.ValueString()
	}

	route.TargetDevices = trafficTargetsFromList(ctx, plan.TargetDevices, diags)
	route.Domains = trafficDomainsFromList(ctx, plan.Domains, diags)

	if !plan.IPAddresses.IsNull() && !plan.IPAddresses.IsUnknown() {
		var ips []string
		diags.Append(plan.IPAddresses.ElementsAs(ctx, &ips, false)...)
		route.IPAddresses = ips
	}

	if !plan.IPRanges.IsNull() && !plan.IPRanges.IsUnknown() {
		var ranges []string
		diags.Append(plan.IPRanges.ElementsAs(ctx, &ranges, false)...)
		route.IPRanges = ranges
	}

	if !plan.Regions.IsNull() && !plan.Regions.IsUnknown() {
		var regions []string
		diags.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
		route.Regions = regions
	}

	return route
}

func (r *TrafficRouteResource) sdkToState(ctx context.Context, route *unifi.TrafficRoute, state *TrafficRouteResourceModel) diag.Diagnostics {
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

	targets, d := trafficTargetsToList(ctx, route.TargetDevices)
	diags.Append(d...)
	state.TargetDevices = targets

	domains, d := trafficDomainsToList(ctx, route.Domains)
	diags.Append(d...)
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
