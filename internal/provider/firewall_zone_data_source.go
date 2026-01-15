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

var _ datasource.DataSource = &FirewallZoneDataSource{}

type FirewallZoneDataSource struct {
	client *AutoLoginClient
}

type FirewallZoneDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ZoneKey    types.String `tfsdk:"zone_key"`
	NetworkIDs types.Set    `tfsdk:"network_ids"`
}

func NewFirewallZoneDataSource() datasource.DataSource {
	return &FirewallZoneDataSource{}
}

func (d *FirewallZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone"
}

func (d *FirewallZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi firewall zone. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall zone. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall zone. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"zone_key": schema.StringAttribute{
				Description: "The zone key for built-in zones (internal, external, gateway, vpn, hotspot, dmz). Empty for custom zones.",
				Computed:    true,
			},
			"network_ids": schema.SetAttribute{
				Description: "Set of network IDs assigned to this zone.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *FirewallZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FirewallZoneDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a firewall zone.",
		)
		return
	}

	var zone *unifi.FirewallZone
	var err error

	if hasID {
		zone, err = d.client.GetFirewallZone(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "firewall zone")
			return
		}
	} else {
		zones, err := d.client.ListFirewallZones(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "firewall zones")
			return
		}

		searchName := config.Name.ValueString()
		for i := range zones {
			if zones[i].Name == searchName {
				zone = &zones[i]
				break
			}
		}

		if zone == nil {
			resp.Diagnostics.AddError(
				"Firewall Zone Not Found",
				fmt.Sprintf("No firewall zone found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, zone, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *FirewallZoneDataSource) sdkToState(ctx context.Context, zone *unifi.FirewallZone, state *FirewallZoneDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(zone.ID)
	state.Name = types.StringValue(zone.Name)

	if zone.ZoneKey != nil && *zone.ZoneKey != "" {
		state.ZoneKey = types.StringValue(*zone.ZoneKey)
	} else {
		state.ZoneKey = types.StringNull()
	}

	if len(zone.NetworkIDs) > 0 {
		networkIDsSet, d := types.SetValueFrom(ctx, types.StringType, zone.NetworkIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.NetworkIDs = networkIDsSet
	} else {
		state.NetworkIDs = types.SetNull(types.StringType)
	}

	return diags
}
