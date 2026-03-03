package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ContentFilteringDataSource{}

type ContentFilteringDataSource struct {
	client *AutoLoginClient
}

type ContentFilteringDataSourceModel struct {
	Enabled           types.Bool `tfsdk:"enabled"`
	BlockedCategories types.Set  `tfsdk:"blocked_categories"`
	AllowedDomains    types.Set  `tfsdk:"allowed_domains"`
	BlockedDomains    types.Set  `tfsdk:"blocked_domains"`
}

func NewContentFilteringDataSource() datasource.DataSource {
	return &ContentFilteringDataSource{}
}

func (d *ContentFilteringDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_filtering"
}

func (d *ContentFilteringDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the current content filtering configuration for the site. Read-only.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{Computed: true},
			"blocked_categories": schema.SetAttribute{
				Description: "Set of blocked content categories.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"allowed_domains": schema.SetAttribute{
				Description: "Set of allowed domains (bypass filtering).",
				Computed:    true,
				ElementType: types.StringType,
			},
			"blocked_domains": schema.SetAttribute{
				Description: "Set of explicitly blocked domains.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *ContentFilteringDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ContentFilteringDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	filtering, err := d.client.GetContentFiltering(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "content filtering")
		return
	}

	var state ContentFilteringDataSourceModel
	state.Enabled = types.BoolValue(derefBool(filtering.Enabled))

	if len(filtering.BlockedCategories) > 0 {
		s, diags := types.SetValueFrom(ctx, types.StringType, filtering.BlockedCategories)
		resp.Diagnostics.Append(diags...)
		state.BlockedCategories = s
	} else {
		state.BlockedCategories = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if len(filtering.AllowedDomains) > 0 {
		s, diags := types.SetValueFrom(ctx, types.StringType, filtering.AllowedDomains)
		resp.Diagnostics.Append(diags...)
		state.AllowedDomains = s
	} else {
		state.AllowedDomains = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if len(filtering.BlockedDomains) > 0 {
		s, diags := types.SetValueFrom(ctx, types.StringType, filtering.BlockedDomains)
		resp.Diagnostics.Append(diags...)
		state.BlockedDomains = s
	} else {
		state.BlockedDomains = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
