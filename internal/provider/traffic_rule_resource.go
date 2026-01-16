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
	_ resource.Resource                = &TrafficRuleResource{}
	_ resource.ResourceWithImportState = &TrafficRuleResource{}
)

type TrafficRuleResource struct {
	client *AutoLoginClient
}

type TrafficRuleResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	Enabled        types.Bool     `tfsdk:"enabled"`
	Action         types.String   `tfsdk:"action"`
	MatchingTarget types.String   `tfsdk:"matching_target"`
	TargetDevices  types.List     `tfsdk:"target_devices"`
	Schedule       types.Object   `tfsdk:"schedule"`
	Description    types.String   `tfsdk:"description"`
	AppCategoryIDs types.Set      `tfsdk:"app_category_ids"`
	AppIDs         types.Set      `tfsdk:"app_ids"`
	Domains        types.List     `tfsdk:"domains"`
	IPAddresses    types.Set      `tfsdk:"ip_addresses"`
	IPRanges       types.Set      `tfsdk:"ip_ranges"`
	Regions        types.Set      `tfsdk:"regions"`
	NetworkID      types.String   `tfsdk:"network_id"`
	BandwidthLimit types.Object   `tfsdk:"bandwidth_limit"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func NewTrafficRuleResource() resource.Resource {
	return &TrafficRuleResource{}
}

func (r *TrafficRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_rule"
}

func (r *TrafficRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi traffic rule for QoS and traffic management.",
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
				Description: "The unique identifier of the traffic rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the traffic rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the traffic rule is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"action": schema.StringAttribute{
				Description: "The action for the rule. Valid values: BLOCK, ALLOW.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("BLOCK", "ALLOW"),
				},
			},
			"matching_target": schema.StringAttribute{
				Description: "The matching target type. Valid values: INTERNET, IP, DOMAIN, REGION, APP.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("INTERNET", "IP", "DOMAIN", "REGION", "APP"),
				},
			},
			"target_devices": schema.ListNestedAttribute{
				Description: "List of target devices for the rule.",
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
			"schedule": schema.SingleNestedAttribute{
				Description: "Schedule for when the rule is active.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Schedule mode. Valid values: ALWAYS, CUSTOM.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALWAYS", "CUSTOM"),
						},
					},
					"time_range_start": schema.StringAttribute{
						Description: "Start time in HH:MM format.",
						Optional:    true,
					},
					"time_range_end": schema.StringAttribute{
						Description: "End time in HH:MM format.",
						Optional:    true,
					},
					"days_of_week": schema.SetAttribute{
						Description: "Days of week when the rule is active. Valid values: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for the traffic rule.",
				Optional:    true,
			},
			"app_category_ids": schema.SetAttribute{
				Description: "Set of application category IDs for app-based filtering.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"app_ids": schema.SetAttribute{
				Description: "Set of application IDs for app-based filtering.",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"domains": schema.ListNestedAttribute{
				Description: "List of domains for domain-based filtering.",
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
				Description: "Set of IP addresses or CIDR blocks for IP-based filtering.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ip_ranges": schema.SetAttribute{
				Description: "Set of IP ranges for IP-based filtering.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"regions": schema.SetAttribute{
				Description: "Set of geographic regions for region-based filtering.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID to apply the rule to.",
				Optional:    true,
			},
			"bandwidth_limit": schema.SingleNestedAttribute{
				Description: "Bandwidth limit configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"download_limit_kbps": schema.Int64Attribute{
						Description: "Download speed limit in Kbps.",
						Optional:    true,
					},
					"upload_limit_kbps": schema.Int64Attribute{
						Description: "Upload speed limit in Kbps.",
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether bandwidth limiting is enabled.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *TrafficRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TrafficRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TrafficRuleResourceModel

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

	rule := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTrafficRule(ctx, rule)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "traffic rule")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrafficRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TrafficRuleResourceModel

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

	rule, err := r.client.GetTrafficRule(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "traffic rule")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, rule, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TrafficRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TrafficRuleResourceModel

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

	var state TrafficRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	rule.ID = state.ID.ValueString()

	updated, err := r.client.UpdateTrafficRule(ctx, state.ID.ValueString(), rule)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "traffic rule")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrafficRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TrafficRuleResourceModel

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

	err := r.client.DeleteTrafficRule(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "traffic rule")
		return
	}
}

func (r *TrafficRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TrafficRuleResource) planToSDK(ctx context.Context, plan *TrafficRuleResourceModel, diags *diag.Diagnostics) *unifi.TrafficRule {
	rule := &unifi.TrafficRule{
		Name:    plan.Name.ValueString(),
		Enabled: boolPtr(plan.Enabled.ValueBool()),
		Action:  plan.Action.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		rule.Description = plan.Description.ValueString()
	}

	if !plan.MatchingTarget.IsNull() && !plan.MatchingTarget.IsUnknown() {
		rule.MatchingTarget = plan.MatchingTarget.ValueString()
	}

	if !plan.NetworkID.IsNull() && !plan.NetworkID.IsUnknown() {
		rule.NetworkID = plan.NetworkID.ValueString()
	}

	rule.TargetDevices = trafficTargetsFromList(ctx, plan.TargetDevices, diags)
	rule.Domains = trafficDomainsFromList(ctx, plan.Domains, diags)
	rule.Schedule = trafficScheduleFromObject(ctx, plan.Schedule, diags)
	rule.BandwidthLimit = trafficBandwidthFromObject(ctx, plan.BandwidthLimit, diags)

	if !plan.AppCategoryIDs.IsNull() && !plan.AppCategoryIDs.IsUnknown() {
		var categories []string
		diags.Append(plan.AppCategoryIDs.ElementsAs(ctx, &categories, false)...)
		rule.AppCategoryIDs = categories
	}

	if !plan.AppIDs.IsNull() && !plan.AppIDs.IsUnknown() {
		var ids []int64
		diags.Append(plan.AppIDs.ElementsAs(ctx, &ids, false)...)
		for _, id := range ids {
			rule.AppIDs = append(rule.AppIDs, int(id))
		}
	}

	if !plan.IPAddresses.IsNull() && !plan.IPAddresses.IsUnknown() {
		var ips []string
		diags.Append(plan.IPAddresses.ElementsAs(ctx, &ips, false)...)
		rule.IPAddresses = ips
	}

	if !plan.IPRanges.IsNull() && !plan.IPRanges.IsUnknown() {
		var ranges []string
		diags.Append(plan.IPRanges.ElementsAs(ctx, &ranges, false)...)
		rule.IPRanges = ranges
	}

	if !plan.Regions.IsNull() && !plan.Regions.IsUnknown() {
		var regions []string
		diags.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
		rule.Regions = regions
	}

	return rule
}

func (r *TrafficRuleResource) sdkToState(ctx context.Context, rule *unifi.TrafficRule, state *TrafficRuleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(rule.ID)
	state.Name = types.StringValue(rule.Name)
	state.Enabled = types.BoolValue(derefBool(rule.Enabled))
	state.Action = types.StringValue(rule.Action)

	if rule.Description != "" {
		state.Description = types.StringValue(rule.Description)
	} else {
		state.Description = types.StringNull()
	}

	if rule.MatchingTarget != "" {
		state.MatchingTarget = types.StringValue(rule.MatchingTarget)
	} else {
		state.MatchingTarget = types.StringNull()
	}

	if rule.NetworkID != "" {
		state.NetworkID = types.StringValue(rule.NetworkID)
	} else {
		state.NetworkID = types.StringNull()
	}

	targets, d := trafficTargetsToList(ctx, rule.TargetDevices)
	diags.Append(d...)
	state.TargetDevices = targets

	domains, d := trafficDomainsToList(ctx, rule.Domains)
	diags.Append(d...)
	state.Domains = domains

	schedule, d := trafficScheduleToObject(ctx, rule.Schedule)
	diags.Append(d...)
	state.Schedule = schedule

	bandwidth, d := trafficBandwidthToObject(ctx, rule.BandwidthLimit)
	diags.Append(d...)
	state.BandwidthLimit = bandwidth

	if len(rule.AppCategoryIDs) > 0 {
		categories, d := types.SetValueFrom(ctx, types.StringType, rule.AppCategoryIDs)
		diags.Append(d...)
		state.AppCategoryIDs = categories
	} else {
		state.AppCategoryIDs = types.SetNull(types.StringType)
	}

	if len(rule.AppIDs) > 0 {
		var ids []int64
		for _, id := range rule.AppIDs {
			ids = append(ids, int64(id))
		}
		appIDs, d := types.SetValueFrom(ctx, types.Int64Type, ids)
		diags.Append(d...)
		state.AppIDs = appIDs
	} else {
		state.AppIDs = types.SetNull(types.Int64Type)
	}

	if len(rule.IPAddresses) > 0 {
		ips, d := types.SetValueFrom(ctx, types.StringType, rule.IPAddresses)
		diags.Append(d...)
		state.IPAddresses = ips
	} else {
		state.IPAddresses = types.SetNull(types.StringType)
	}

	if len(rule.IPRanges) > 0 {
		ranges, d := types.SetValueFrom(ctx, types.StringType, rule.IPRanges)
		diags.Append(d...)
		state.IPRanges = ranges
	} else {
		state.IPRanges = types.SetNull(types.StringType)
	}

	if len(rule.Regions) > 0 {
		regions, d := types.SetValueFrom(ctx, types.StringType, rule.Regions)
		diags.Append(d...)
		state.Regions = regions
	} else {
		state.Regions = types.SetNull(types.StringType)
	}

	return diags
}
