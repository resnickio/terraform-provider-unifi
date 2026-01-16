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

var _ datasource.DataSource = &FirewallRuleDataSource{}

type FirewallRuleDataSource struct {
	client *AutoLoginClient
}

type FirewallRuleDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SiteID              types.String `tfsdk:"site_id"`
	Name                types.String `tfsdk:"name"`
	Ruleset             types.String `tfsdk:"ruleset"`
	Action              types.String `tfsdk:"action"`
	RuleIndex           types.Int64  `tfsdk:"rule_index"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Protocol            types.String `tfsdk:"protocol"`
	SrcNetworkConfType  types.String `tfsdk:"src_network_conf_type"`
	SrcAddress          types.String `tfsdk:"src_address"`
	SrcFirewallGroupIDs types.Set    `tfsdk:"src_firewall_group_ids"`
	DstNetworkConfType  types.String `tfsdk:"dst_network_conf_type"`
	DstAddress          types.String `tfsdk:"dst_address"`
	DstFirewallGroupIDs types.Set    `tfsdk:"dst_firewall_group_ids"`
	DstPort             types.String `tfsdk:"dst_port"`
	Logging             types.Bool   `tfsdk:"logging"`
	StateNew            types.Bool   `tfsdk:"state_new"`
	StateEstablished    types.Bool   `tfsdk:"state_established"`
	StateRelated        types.Bool   `tfsdk:"state_related"`
	StateInvalid        types.Bool   `tfsdk:"state_invalid"`
}

func NewFirewallRuleDataSource() datasource.DataSource {
	return &FirewallRuleDataSource{}
}

func (d *FirewallRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (d *FirewallRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi legacy firewall rule. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the firewall rule exists.",
				Computed:    true,
			},
			"ruleset": schema.StringAttribute{
				Description: "The ruleset this rule belongs to.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "The action to take (accept, drop, reject).",
				Computed:    true,
			},
			"rule_index": schema.Int64Attribute{
				Description: "The index/priority of the rule.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled.",
				Computed:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol to match.",
				Computed:    true,
			},
			"src_network_conf_type": schema.StringAttribute{
				Description: "Source network configuration type.",
				Computed:    true,
			},
			"src_address": schema.StringAttribute{
				Description: "Source IP address or CIDR.",
				Computed:    true,
			},
			"src_firewall_group_ids": schema.SetAttribute{
				Description: "Set of source firewall group IDs.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"dst_network_conf_type": schema.StringAttribute{
				Description: "Destination network configuration type.",
				Computed:    true,
			},
			"dst_address": schema.StringAttribute{
				Description: "Destination IP address or CIDR.",
				Computed:    true,
			},
			"dst_firewall_group_ids": schema.SetAttribute{
				Description: "Set of destination firewall group IDs.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"dst_port": schema.StringAttribute{
				Description: "Destination port or port range.",
				Computed:    true,
			},
			"logging": schema.BoolAttribute{
				Description: "Whether to log matching packets.",
				Computed:    true,
			},
			"state_new": schema.BoolAttribute{
				Description: "Match new connections.",
				Computed:    true,
			},
			"state_established": schema.BoolAttribute{
				Description: "Match established connections.",
				Computed:    true,
			},
			"state_related": schema.BoolAttribute{
				Description: "Match related connections.",
				Computed:    true,
			},
			"state_invalid": schema.BoolAttribute{
				Description: "Match invalid connections.",
				Computed:    true,
			},
		},
	}
}

func (d *FirewallRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FirewallRuleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a firewall rule.",
		)
		return
	}

	var rule *unifi.FirewallRule
	var err error

	if hasID {
		rule, err = d.client.GetFirewallRule(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "firewall rule")
			return
		}
	} else {
		rules, err := d.client.ListFirewallRules(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "firewall rules")
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
				"Firewall Rule Not Found",
				fmt.Sprintf("No firewall rule found with name '%s'.", searchName),
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

func (d *FirewallRuleDataSource) sdkToState(ctx context.Context, rule *unifi.FirewallRule, state *FirewallRuleDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(rule.ID)
	state.SiteID = types.StringValue(rule.SiteID)
	state.Name = types.StringValue(rule.Name)
	state.Ruleset = types.StringValue(rule.Ruleset)
	state.Action = types.StringValue(rule.Action)

	if rule.RuleIndex != nil {
		state.RuleIndex = types.Int64Value(int64(*rule.RuleIndex))
	} else {
		state.RuleIndex = types.Int64Null()
	}

	state.Enabled = types.BoolValue(derefBool(rule.Enabled))
	state.Protocol = types.StringValue(rule.Protocol)

	if rule.SrcNetworkConfType != "" {
		state.SrcNetworkConfType = types.StringValue(rule.SrcNetworkConfType)
	} else {
		state.SrcNetworkConfType = types.StringNull()
	}

	if rule.SrcAddress != "" {
		state.SrcAddress = types.StringValue(rule.SrcAddress)
	} else {
		state.SrcAddress = types.StringNull()
	}

	if len(rule.SrcFirewallGroupIDs) > 0 {
		srcSet, d := types.SetValueFrom(ctx, types.StringType, rule.SrcFirewallGroupIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.SrcFirewallGroupIDs = srcSet
	} else {
		state.SrcFirewallGroupIDs = types.SetNull(types.StringType)
	}

	if rule.DstNetworkConfType != "" {
		state.DstNetworkConfType = types.StringValue(rule.DstNetworkConfType)
	} else {
		state.DstNetworkConfType = types.StringNull()
	}

	if rule.DstAddress != "" {
		state.DstAddress = types.StringValue(rule.DstAddress)
	} else {
		state.DstAddress = types.StringNull()
	}

	if len(rule.DstFirewallGroupIDs) > 0 {
		dstSet, d := types.SetValueFrom(ctx, types.StringType, rule.DstFirewallGroupIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.DstFirewallGroupIDs = dstSet
	} else {
		state.DstFirewallGroupIDs = types.SetNull(types.StringType)
	}

	if rule.DstPort != "" {
		state.DstPort = types.StringValue(rule.DstPort)
	} else {
		state.DstPort = types.StringNull()
	}

	state.Logging = types.BoolValue(derefBool(rule.Logging))
	state.StateNew = types.BoolValue(derefBool(rule.StateNew))
	state.StateEstablished = types.BoolValue(derefBool(rule.StateEstablished))
	state.StateRelated = types.BoolValue(derefBool(rule.StateRelated))
	state.StateInvalid = types.BoolValue(derefBool(rule.StateInvalid))

	return diags
}
