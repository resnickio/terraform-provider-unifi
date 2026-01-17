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

var _ datasource.DataSource = &RADIUSProfileDataSource{}

type RADIUSProfileDataSource struct {
	client *AutoLoginClient
}

type RADIUSProfileDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SiteID                types.String `tfsdk:"site_id"`
	Name                  types.String `tfsdk:"name"`
	UseUsgAuthServer      types.Bool   `tfsdk:"use_usg_auth_server"`
	UseUsgAcctServer      types.Bool   `tfsdk:"use_usg_acct_server"`
	VlanEnabled           types.Bool   `tfsdk:"vlan_enabled"`
	VlanWlanMode          types.String `tfsdk:"vlan_wlan_mode"`
	InterimUpdateEnabled  types.Bool   `tfsdk:"interim_update_enabled"`
	InterimUpdateInterval types.Int64  `tfsdk:"interim_update_interval"`
}

func NewRADIUSProfileDataSource() datasource.DataSource {
	return &RADIUSProfileDataSource{}
}

func (d *RADIUSProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radius_profile"
}

func (d *RADIUSProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi RADIUS profile. Lookup by either id or name. Note: auth_server and acct_server are not exposed as they contain write-only secret fields.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the RADIUS profile. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the RADIUS profile. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the RADIUS profile exists.",
				Computed:    true,
			},
			"use_usg_auth_server": schema.BoolAttribute{
				Description: "Use the USG/UDM as the authentication server.",
				Computed:    true,
			},
			"use_usg_acct_server": schema.BoolAttribute{
				Description: "Use the USG/UDM as the accounting server.",
				Computed:    true,
			},
			"vlan_enabled": schema.BoolAttribute{
				Description: "Enable VLAN assignment via RADIUS.",
				Computed:    true,
			},
			"vlan_wlan_mode": schema.StringAttribute{
				Description: "VLAN WLAN mode. Valid values: disabled, optional, required.",
				Computed:    true,
			},
			"interim_update_enabled": schema.BoolAttribute{
				Description: "Enable interim accounting updates.",
				Computed:    true,
			},
			"interim_update_interval": schema.Int64Attribute{
				Description: "Interval in seconds between interim accounting updates.",
				Computed:    true,
			},
		},
	}
}

func (d *RADIUSProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RADIUSProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RADIUSProfileDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a RADIUS profile.",
		)
		return
	}

	var profile *unifi.RADIUSProfile
	var err error

	if hasID {
		profile, err = d.client.GetRADIUSProfile(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "RADIUS profile")
			return
		}
	} else {
		profiles, err := d.client.ListRADIUSProfiles(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "RADIUS profiles")
			return
		}

		searchName := config.Name.ValueString()
		for i := range profiles {
			if profiles[i].Name == searchName {
				profile = &profiles[i]
				break
			}
		}

		if profile == nil {
			resp.Diagnostics.AddError(
				"RADIUS Profile Not Found",
				fmt.Sprintf("No RADIUS profile found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, profile, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *RADIUSProfileDataSource) sdkToState(ctx context.Context, profile *unifi.RADIUSProfile, state *RADIUSProfileDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(profile.ID)
	state.SiteID = stringValueOrNull(profile.SiteID)
	state.Name = types.StringValue(profile.Name)
	state.UseUsgAuthServer = types.BoolValue(derefBool(profile.UseUsgAuthServer))
	state.UseUsgAcctServer = types.BoolValue(derefBool(profile.UseUsgAcctServer))
	state.VlanEnabled = types.BoolValue(derefBool(profile.VlanEnabled))
	state.InterimUpdateEnabled = types.BoolValue(derefBool(profile.InterimUpdateEnabled))

	if profile.VlanWlanMode != "" {
		state.VlanWlanMode = types.StringValue(profile.VlanWlanMode)
	} else {
		state.VlanWlanMode = types.StringNull()
	}

	if profile.InterimUpdateInterval != nil {
		state.InterimUpdateInterval = types.Int64Value(int64(*profile.InterimUpdateInterval))
	} else {
		state.InterimUpdateInterval = types.Int64Null()
	}

	return diags
}
