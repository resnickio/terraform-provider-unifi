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

var _ datasource.DataSource = &NatRuleDataSource{}

type NatRuleDataSource struct {
	client *AutoLoginClient
}

type NatRuleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Protocol    types.String `tfsdk:"protocol"`
	Logging     types.Bool   `tfsdk:"logging"`
}

func NewNatRuleDataSource() datasource.DataSource {
	return &NatRuleDataSource{}
}

func (d *NatRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_rule"
}

func (d *NatRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi NAT rule. Lookup by either id or description.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the NAT rule. Specify either id or description.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("description")),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the NAT rule. Specify either id or description.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the NAT rule is enabled.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The NAT rule type (MASQUERADE, DNAT, SNAT).",
				Computed:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol for the NAT rule.",
				Computed:    true,
			},
			"logging": schema.BoolAttribute{
				Description: "Whether to log traffic matching this rule.",
				Computed:    true,
			},
		},
	}
}

func (d *NatRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NatRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NatRuleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasDescription := !config.Description.IsNull() && config.Description.ValueString() != ""

	if !hasID && !hasDescription {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'description' must be specified to look up a NAT rule.",
		)
		return
	}

	var rule *unifi.NatRule
	var err error

	if hasID {
		rule, err = d.client.GetNatRule(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "NAT rule")
			return
		}
	} else {
		rules, err := d.client.ListNatRules(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "NAT rules")
			return
		}

		searchDescription := config.Description.ValueString()
		for i := range rules {
			if rules[i].Description == searchDescription {
				rule = &rules[i]
				break
			}
		}

		if rule == nil {
			resp.Diagnostics.AddError(
				"NAT Rule Not Found",
				fmt.Sprintf("No NAT rule found with description '%s'.", searchDescription),
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

func (d *NatRuleDataSource) sdkToState(ctx context.Context, rule *unifi.NatRule, state *NatRuleDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(rule.ID)
	state.Enabled = types.BoolValue(derefBool(rule.Enabled))
	state.Type = types.StringValue(rule.Type)
	state.Protocol = stringValueOrNull(rule.Protocol)
	state.Logging = types.BoolValue(derefBool(rule.Logging))

	if rule.Description != "" {
		state.Description = types.StringValue(rule.Description)
	} else {
		state.Description = types.StringNull()
	}

	return diags
}
