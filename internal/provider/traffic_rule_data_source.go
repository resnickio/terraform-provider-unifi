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

var _ datasource.DataSource = &TrafficRuleDataSource{}

type TrafficRuleDataSource struct {
	client *AutoLoginClient
}

type TrafficRuleDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Action         types.String `tfsdk:"action"`
	MatchingTarget types.String `tfsdk:"matching_target"`
	TargetDevices  types.List   `tfsdk:"target_devices"`
	Schedule       types.Object `tfsdk:"schedule"`
	Description    types.String `tfsdk:"description"`
	AppCategoryIDs types.Set    `tfsdk:"app_category_ids"`
	AppIDs         types.Set    `tfsdk:"app_ids"`
	Domains        types.List   `tfsdk:"domains"`
	IPAddresses    types.Set    `tfsdk:"ip_addresses"`
	IPRanges       types.Set    `tfsdk:"ip_ranges"`
	Regions        types.Set    `tfsdk:"regions"`
	NetworkID      types.String `tfsdk:"network_id"`
	BandwidthLimit types.Object `tfsdk:"bandwidth_limit"`
}

func NewTrafficRuleDataSource() datasource.DataSource {
	return &TrafficRuleDataSource{}
}

func (d *TrafficRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_rule"
}

func (d *TrafficRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi traffic rule (QoS/blocking). Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the traffic rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the traffic rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the traffic rule is enabled.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "The action for the rule (BLOCK, ALLOW).",
				Computed:    true,
			},
			"matching_target": schema.StringAttribute{
				Description: "The matching target type (INTERNET, IP, DOMAIN, REGION, APP).",
				Computed:    true,
			},
			"target_devices": schema.ListNestedAttribute{
				Description: "List of target devices for the rule.",
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
			"schedule": schema.SingleNestedAttribute{
				Description: "Schedule for when the rule is active.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Schedule mode (ALWAYS, CUSTOM).",
						Computed:    true,
					},
					"time_range_start": schema.StringAttribute{
						Description: "Start time in HH:MM format.",
						Computed:    true,
					},
					"time_range_end": schema.StringAttribute{
						Description: "End time in HH:MM format.",
						Computed:    true,
					},
					"days_of_week": schema.SetAttribute{
						Description: "Days of week when the rule is active.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for the traffic rule.",
				Computed:    true,
			},
			"app_category_ids": schema.SetAttribute{
				Description: "Set of application category IDs for app-based filtering.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"app_ids": schema.SetAttribute{
				Description: "Set of application IDs for app-based filtering.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"ip_addresses": schema.SetAttribute{
				Description: "Set of IP addresses or CIDR blocks for IP-based filtering.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"ip_ranges": schema.SetAttribute{
				Description: "Set of IP ranges for IP-based filtering.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"regions": schema.SetAttribute{
				Description: "Set of geographic regions for region-based filtering.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID to apply the rule to.",
				Computed:    true,
			},
			"domains": schema.ListNestedAttribute{
				Description: "List of domains for domain-based filtering.",
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
			"bandwidth_limit": schema.SingleNestedAttribute{
				Description: "Bandwidth limit configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"download_limit_kbps": schema.Int64Attribute{
						Description: "Download speed limit in Kbps.",
						Computed:    true,
					},
					"upload_limit_kbps": schema.Int64Attribute{
						Description: "Upload speed limit in Kbps.",
						Computed:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether bandwidth limiting is enabled.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (d *TrafficRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TrafficRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config TrafficRuleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a traffic rule.",
		)
		return
	}

	var rule *unifi.TrafficRule
	var err error

	if hasID {
		rule, err = d.client.GetTrafficRule(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "traffic rule")
			return
		}
	} else {
		rules, err := d.client.ListTrafficRules(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "traffic rules")
			return
		}

		searchName := config.Name.ValueString()
		for i := range rules {
			if rules[i].Name == searchName {
				rule = &rules[i]
				break
			}
		}

		if rule == nil {
			resp.Diagnostics.AddError(
				"Traffic Rule Not Found",
				fmt.Sprintf("No traffic rule found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, rule, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *TrafficRuleDataSource) sdkToState(ctx context.Context, rule *unifi.TrafficRule, state *TrafficRuleDataSourceModel) diag.Diagnostics {
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

	targets, diagsTargets := trafficTargetsToList(ctx, rule.TargetDevices)
	diags.Append(diagsTargets...)
	state.TargetDevices = targets

	domains, diagsDomains := trafficDomainsToList(ctx, rule.Domains)
	diags.Append(diagsDomains...)
	state.Domains = domains

	schedule, diagsSchedule := trafficScheduleToObject(ctx, rule.Schedule)
	diags.Append(diagsSchedule...)
	state.Schedule = schedule

	bandwidth, diagsBandwidth := trafficBandwidthToObject(ctx, rule.BandwidthLimit)
	diags.Append(diagsBandwidth...)
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
