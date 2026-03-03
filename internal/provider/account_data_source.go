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

var _ datasource.DataSource = &AccountDataSource{}

type AccountDataSource struct {
	client *AutoLoginClient
}

type AccountDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	SiteID           types.String `tfsdk:"site_id"`
	Name             types.String `tfsdk:"name"`
	TunnelConfigType types.String `tfsdk:"tunnel_config_type"`
	TunnelMediumType types.Int64  `tfsdk:"tunnel_medium_type"`
	TunnelType       types.Int64  `tfsdk:"tunnel_type"`
	VLAN             types.Int64  `tfsdk:"vlan"`
	NetworkConfID    types.String `tfsdk:"network_id"`
}

func NewAccountDataSource() datasource.DataSource {
	return &AccountDataSource{}
}

func (d *AccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *AccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi RADIUS account. Lookup by id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the RADIUS account. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The username of the RADIUS account. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Computed: true,
			},
			"tunnel_config_type": schema.StringAttribute{
				Computed: true,
			},
			"tunnel_medium_type": schema.Int64Attribute{
				Computed: true,
			},
			"tunnel_type": schema.Int64Attribute{
				Computed: true,
			},
			"vlan": schema.Int64Attribute{
				Computed: true,
			},
			"network_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *AccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a RADIUS account.",
		)
		return
	}

	var account *unifi.RADIUSAccount

	if hasID {
		var err error
		account, err = d.client.GetRADIUSAccount(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "RADIUS account")
			return
		}
	} else {
		accounts, err := d.client.ListRADIUSAccounts(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "RADIUS accounts")
			return
		}
		searchName := config.Name.ValueString()
		for i := range accounts {
			if accounts[i].Name == searchName {
				account = &accounts[i]
				break
			}
		}
		if account == nil {
			resp.Diagnostics.AddError(
				"RADIUS Account Not Found",
				fmt.Sprintf("No RADIUS account found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(account, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *AccountDataSource) sdkToState(account *unifi.RADIUSAccount, state *AccountDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(account.ID)
	state.SiteID = types.StringValue(account.SiteID)
	state.Name = types.StringValue(account.Name)
	state.TunnelConfigType = stringValueOrNull(account.TunnelConfigType)
	state.NetworkConfID = stringValueOrNull(account.NetworkConfID)

	if account.TunnelMediumType != nil {
		state.TunnelMediumType = types.Int64Value(int64(*account.TunnelMediumType))
	} else {
		state.TunnelMediumType = types.Int64Null()
	}

	if account.TunnelType != nil {
		state.TunnelType = types.Int64Value(int64(*account.TunnelType))
	} else {
		state.TunnelType = types.Int64Null()
	}

	if account.VLAN != nil {
		state.VLAN = types.Int64Value(int64(*account.VLAN))
	} else {
		state.VLAN = types.Int64Null()
	}

	return diags
}
