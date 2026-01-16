package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ datasource.DataSource = &FirewallPolicyDataSource{}

type FirewallPolicyDataSource struct {
	client *AutoLoginClient
}

type FirewallPolicyDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Action              types.String `tfsdk:"action"`
	Protocol            types.String `tfsdk:"protocol"`
	IPVersion           types.String `tfsdk:"ip_version"`
	Index               types.Int64  `tfsdk:"index"`
	Logging             types.Bool   `tfsdk:"logging"`
	ConnectionStateType types.String `tfsdk:"connection_state_type"`
	ConnectionStates    types.Set    `tfsdk:"connection_states"`
	MatchIPSec          types.Bool   `tfsdk:"match_ipsec"`
	ICMPTypename        types.String `tfsdk:"icmp_typename"`
	ICMPV6Typename      types.String `tfsdk:"icmpv6_typename"`
	Source              types.Object `tfsdk:"source"`
	Destination         types.Object `tfsdk:"destination"`
	Schedule            types.Object `tfsdk:"schedule"`
}

func NewFirewallPolicyDataSource() datasource.DataSource {
	return &FirewallPolicyDataSource{}
}

func (d *FirewallPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_policy"
}

func (d *FirewallPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi firewall policy (v2 zone-based firewall). Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall policy. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall policy. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the policy is enabled.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "The action to take (ALLOW, BLOCK, REJECT).",
				Computed:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol to match.",
				Computed:    true,
			},
			"ip_version": schema.StringAttribute{
				Description: "The IP version to match (BOTH, IPV4, IPV6).",
				Computed:    true,
			},
			"index": schema.Int64Attribute{
				Description: "The index/priority of the policy.",
				Computed:    true,
			},
			"logging": schema.BoolAttribute{
				Description: "Whether to log matching packets.",
				Computed:    true,
			},
			"connection_state_type": schema.StringAttribute{
				Description: "Connection state matching type (ALL, RESPOND_ONLY, CUSTOM).",
				Computed:    true,
			},
			"connection_states": schema.SetAttribute{
				Description: "Set of connection states to match.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"match_ipsec": schema.BoolAttribute{
				Description: "Whether to match IPSec traffic.",
				Computed:    true,
			},
			"icmp_typename": schema.StringAttribute{
				Description: "ICMP type name (for ICMP protocol).",
				Computed:    true,
			},
			"icmpv6_typename": schema.StringAttribute{
				Description: "ICMPv6 type name (for ICMPv6 protocol).",
				Computed:    true,
			},
			"source": schema.SingleNestedAttribute{
				Description: "Source matching criteria.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"zone_id": schema.StringAttribute{
						Description: "The source zone ID.",
						Computed:    true,
					},
					"matching_target": schema.StringAttribute{
						Description: "Matching target type.",
						Computed:    true,
					},
					"ips": schema.SetAttribute{
						Description: "Set of IP addresses or CIDR ranges to match.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"mac": schema.StringAttribute{
						Description: "MAC address to match.",
						Computed:    true,
					},
					"port": schema.StringAttribute{
						Description: "Port or port range to match.",
						Computed:    true,
					},
					"network_id": schema.StringAttribute{
						Description: "Network ID to match.",
						Computed:    true,
					},
					"client_macs": schema.SetAttribute{
						Description: "Set of client MAC addresses to match.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"destination": schema.SingleNestedAttribute{
				Description: "Destination matching criteria.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"zone_id": schema.StringAttribute{
						Description: "The destination zone ID.",
						Computed:    true,
					},
					"matching_target": schema.StringAttribute{
						Description: "Matching target type.",
						Computed:    true,
					},
					"ips": schema.SetAttribute{
						Description: "Set of IP addresses or CIDR ranges to match.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"mac": schema.StringAttribute{
						Description: "MAC address to match.",
						Computed:    true,
					},
					"port": schema.StringAttribute{
						Description: "Port or port range to match.",
						Computed:    true,
					},
					"network_id": schema.StringAttribute{
						Description: "Network ID to match.",
						Computed:    true,
					},
					"client_macs": schema.SetAttribute{
						Description: "Set of client MAC addresses to match.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				Description: "Schedule configuration for when the policy is active.",
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
						Description: "Days of the week.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (d *FirewallPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FirewallPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a firewall policy.",
		)
		return
	}

	var policy *unifi.FirewallPolicy
	var err error

	if hasID {
		policy, err = d.client.GetFirewallPolicy(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "firewall policy")
			return
		}
	} else {
		policies, err := d.client.ListFirewallPolicies(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "firewall policies")
			return
		}

		searchName := config.Name.ValueString()
		for i := range policies {
			if policies[i].Name == searchName {
				policy = &policies[i]
				break
			}
		}

		if policy == nil {
			resp.Diagnostics.AddError(
				"Firewall Policy Not Found",
				fmt.Sprintf("No firewall policy found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, policy, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *FirewallPolicyDataSource) sdkToState(ctx context.Context, policy *unifi.FirewallPolicy, state *FirewallPolicyDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	state.Enabled = types.BoolValue(derefBool(policy.Enabled))
	state.Action = types.StringValue(policy.Action)
	state.Protocol = types.StringValue(policy.Protocol)
	state.IPVersion = types.StringValue(policy.IPVersion)

	if policy.Index != nil {
		state.Index = types.Int64Value(int64(*policy.Index))
	} else {
		state.Index = types.Int64Null()
	}

	state.Logging = types.BoolValue(derefBool(policy.Logging))
	state.ConnectionStateType = types.StringValue(policy.ConnectionStateType)

	if len(policy.ConnectionStates) > 0 {
		statesSet, d := types.SetValueFrom(ctx, types.StringType, policy.ConnectionStates)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.ConnectionStates = statesSet
	} else {
		state.ConnectionStates = types.SetNull(types.StringType)
	}

	state.MatchIPSec = types.BoolValue(derefBool(policy.MatchIPSec))

	if policy.ICMPTypename != "" {
		state.ICMPTypename = types.StringValue(policy.ICMPTypename)
	} else {
		state.ICMPTypename = types.StringNull()
	}

	if policy.ICMPV6Typename != "" {
		state.ICMPV6Typename = types.StringValue(policy.ICMPV6Typename)
	} else {
		state.ICMPV6Typename = types.StringNull()
	}

	if policy.Source != nil {
		sourceObj, d := d.endpointToObject(ctx, policy.Source)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Source = sourceObj
	} else {
		state.Source = types.ObjectNull(endpointAttrTypes)
	}

	if policy.Destination != nil {
		destObj, d := d.endpointToObject(ctx, policy.Destination)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Destination = destObj
	} else {
		state.Destination = types.ObjectNull(endpointAttrTypes)
	}

	if policy.Schedule != nil {
		scheduleObj, d := d.scheduleToObject(ctx, policy.Schedule)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Schedule = scheduleObj
	} else {
		state.Schedule = types.ObjectNull(scheduleAttrTypes)
	}

	return diags
}

func (d *FirewallPolicyDataSource) endpointToObject(ctx context.Context, endpoint *unifi.PolicyEndpoint) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var ipsVal types.Set
	if len(endpoint.IPs) > 0 {
		ipsSet, d := types.SetValueFrom(ctx, types.StringType, endpoint.IPs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(endpointAttrTypes), diags
		}
		ipsVal = ipsSet
	} else {
		ipsVal = types.SetNull(types.StringType)
	}

	var clientMacsVal types.Set
	if len(endpoint.ClientMACs) > 0 {
		macsSet, d := types.SetValueFrom(ctx, types.StringType, endpoint.ClientMACs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(endpointAttrTypes), diags
		}
		clientMacsVal = macsSet
	} else {
		clientMacsVal = types.SetNull(types.StringType)
	}

	attrs := map[string]attr.Value{
		"zone_id":         stringValueOrNull(endpoint.ZoneID),
		"matching_target": stringValueOrNull(endpoint.MatchingTarget),
		"ips":             ipsVal,
		"mac":             stringValueOrNull(endpoint.MAC),
		"port":            stringValueOrNull(endpoint.Port),
		"network_id":      stringValueOrNull(endpoint.NetworkID),
		"client_macs":     clientMacsVal,
	}

	obj, diagObj := types.ObjectValue(endpointAttrTypes, attrs)
	diags.Append(diagObj...)
	return obj, diags
}

func (ds *FirewallPolicyDataSource) scheduleToObject(ctx context.Context, schedule *unifi.PolicySchedule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var daysVal types.Set
	if len(schedule.DaysOfWeek) > 0 {
		daysSet, d := types.SetValueFrom(ctx, types.StringType, schedule.DaysOfWeek)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(scheduleAttrTypes), diags
		}
		daysVal = daysSet
	} else {
		daysVal = types.SetNull(types.StringType)
	}

	attrs := map[string]attr.Value{
		"mode":             stringValueOrNull(schedule.Mode),
		"time_range_start": stringValueOrNull(schedule.TimeRangeStart),
		"time_range_end":   stringValueOrNull(schedule.TimeRangeEnd),
		"days_of_week":     daysVal,
	}

	obj, diagObj := types.ObjectValue(scheduleAttrTypes, attrs)
	diags.Append(diagObj...)
	return obj, diags
}
