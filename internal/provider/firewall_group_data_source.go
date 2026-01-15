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

var _ datasource.DataSource = &FirewallGroupDataSource{}

type FirewallGroupDataSource struct {
	client *AutoLoginClient
}

type FirewallGroupDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	SiteID    types.String `tfsdk:"site_id"`
	Name      types.String `tfsdk:"name"`
	GroupType types.String `tfsdk:"group_type"`
	Members   types.Set    `tfsdk:"members"`
}

func NewFirewallGroupDataSource() datasource.DataSource {
	return &FirewallGroupDataSource{}
}

func (d *FirewallGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_group"
}

func (d *FirewallGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi firewall group. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall group. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall group. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the firewall group exists.",
				Computed:    true,
			},
			"group_type": schema.StringAttribute{
				Description: "The type of the firewall group: 'address-group', 'port-group', or 'ipv6-address-group'.",
				Computed:    true,
			},
			"members": schema.SetAttribute{
				Description: "The members of the firewall group. For address groups, IP addresses or CIDR ranges. For port groups, port numbers or ranges.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *FirewallGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FirewallGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a firewall group.",
		)
		return
	}

	var group *unifi.FirewallGroup
	var err error

	if hasID {
		group, err = d.client.GetFirewallGroup(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "firewall group")
			return
		}
	} else {
		groups, err := d.client.ListFirewallGroups(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "firewall groups")
			return
		}

		searchName := config.Name.ValueString()
		for i := range groups {
			if groups[i].Name == searchName {
				group = &groups[i]
				break
			}
		}

		if group == nil {
			resp.Diagnostics.AddError(
				"Firewall Group Not Found",
				fmt.Sprintf("No firewall group found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, group, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *FirewallGroupDataSource) sdkToState(ctx context.Context, group *unifi.FirewallGroup, state *FirewallGroupDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(group.ID)
	state.SiteID = types.StringValue(group.SiteID)
	state.Name = types.StringValue(group.Name)
	state.GroupType = types.StringValue(group.GroupType)

	if group.GroupMembers == nil {
		group.GroupMembers = []string{}
	}
	members, setDiags := types.SetValueFrom(ctx, types.StringType, group.GroupMembers)
	diags.Append(setDiags...)
	if diags.HasError() {
		return diags
	}
	state.Members = members

	return diags
}
