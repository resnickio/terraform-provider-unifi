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

var _ datasource.DataSource = &SiteDataSource{}

type SiteDataSource struct {
	client *AutoLoginClient
}

type SiteDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewSiteDataSource() datasource.DataSource {
	return &SiteDataSource{}
}

func (d *SiteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (d *SiteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi site. Lookup by id or name (matches against both internal name and description).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the site. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the site (matches internal name or description). Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The human-readable description of the site.",
				Computed:    true,
			},
		},
	}
}

func (d *SiteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *SiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SiteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a site.",
		)
		return
	}

	sites, err := d.client.ListSites(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "sites")
		return
	}

	var site *unifi.NetworkSite

	if hasID {
		searchID := config.ID.ValueString()
		for i := range sites {
			if sites[i].ID == searchID {
				site = &sites[i]
				break
			}
		}
		if site == nil {
			resp.Diagnostics.AddError(
				"Site Not Found",
				fmt.Sprintf("No site found with id '%s'.", searchID),
			)
			return
		}
	} else {
		searchName := config.Name.ValueString()
		for i := range sites {
			if sites[i].Name == searchName || sites[i].Desc == searchName {
				site = &sites[i]
				break
			}
		}
		if site == nil {
			resp.Diagnostics.AddError(
				"Site Not Found",
				fmt.Sprintf("No site found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(site, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *SiteDataSource) sdkToState(site *unifi.NetworkSite, state *SiteDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(site.ID)
	state.Name = types.StringValue(site.Name)
	state.Description = types.StringValue(site.Desc)

	return diags
}
