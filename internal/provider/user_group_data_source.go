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

var _ datasource.DataSource = &UserGroupDataSource{}

type UserGroupDataSource struct {
	client *AutoLoginClient
}

type UserGroupDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	SiteID         types.String `tfsdk:"site_id"`
	Name           types.String `tfsdk:"name"`
	QosRateMaxDown types.Int64  `tfsdk:"qos_rate_max_down"`
	QosRateMaxUp   types.Int64  `tfsdk:"qos_rate_max_up"`
}

func NewUserGroupDataSource() datasource.DataSource {
	return &UserGroupDataSource{}
}

func (d *UserGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (d *UserGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi user group (bandwidth profile). Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user group. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the user group. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the user group exists.",
				Computed:    true,
			},
			"qos_rate_max_down": schema.Int64Attribute{
				Description: "Maximum download rate in kbps. -1 means unlimited.",
				Computed:    true,
			},
			"qos_rate_max_up": schema.Int64Attribute{
				Description: "Maximum upload rate in kbps. -1 means unlimited.",
				Computed:    true,
			},
		},
	}
}

func (d *UserGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a user group.",
		)
		return
	}

	var group *unifi.UserGroup
	var err error

	if hasID {
		group, err = d.client.GetUserGroup(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "user group")
			return
		}
	} else {
		groups, err := d.client.ListUserGroups(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "user groups")
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
				"User Group Not Found",
				fmt.Sprintf("No user group found with name '%s'.", searchName),
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

func (d *UserGroupDataSource) sdkToState(ctx context.Context, group *unifi.UserGroup, state *UserGroupDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(group.ID)
	state.SiteID = types.StringValue(group.SiteID)
	state.Name = types.StringValue(group.Name)

	if group.QosRateMaxDown != nil {
		state.QosRateMaxDown = types.Int64Value(int64(*group.QosRateMaxDown))
	} else {
		state.QosRateMaxDown = types.Int64Null()
	}

	if group.QosRateMaxUp != nil {
		state.QosRateMaxUp = types.Int64Value(int64(*group.QosRateMaxUp))
	} else {
		state.QosRateMaxUp = types.Int64Null()
	}

	return diags
}
