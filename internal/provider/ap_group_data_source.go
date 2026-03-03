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

var _ datasource.DataSource = &APGroupDataSource{}

type APGroupDataSource struct {
	client *AutoLoginClient
}

type APGroupDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DeviceMACs types.Set    `tfsdk:"device_macs"`
}

func NewAPGroupDataSource() datasource.DataSource {
	return &APGroupDataSource{}
}

func (d *APGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ap_group"
}

func (d *APGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi AP group. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the AP group. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the AP group. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"device_macs": schema.SetAttribute{
				Description: "The MAC addresses of devices in this AP group.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *APGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config APGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up an AP group.",
		)
		return
	}

	groups, err := d.client.ListAPGroups(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "AP groups")
		return
	}

	var group *unifi.APGroup

	if hasID {
		searchID := config.ID.ValueString()
		for i := range groups {
			if groups[i].ID == searchID {
				group = &groups[i]
				break
			}
		}
		if group == nil {
			resp.Diagnostics.AddError(
				"AP Group Not Found",
				fmt.Sprintf("No AP group found with id '%s'.", searchID),
			)
			return
		}
	} else {
		searchName := config.Name.ValueString()
		for i := range groups {
			if groups[i].Name == searchName {
				group = &groups[i]
				break
			}
		}
		if group == nil {
			resp.Diagnostics.AddError(
				"AP Group Not Found",
				fmt.Sprintf("No AP group found with name '%s'.", searchName),
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

func (d *APGroupDataSource) sdkToState(ctx context.Context, group *unifi.APGroup, state *APGroupDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(group.ID)
	state.Name = types.StringValue(group.Name)

	if len(group.DeviceMACs) > 0 {
		macsSet, d := types.SetValueFrom(ctx, types.StringType, group.DeviceMACs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.DeviceMACs = macsSet
	} else {
		state.DeviceMACs = types.SetValueMust(types.StringType, []attr.Value{})
	}

	return diags
}
