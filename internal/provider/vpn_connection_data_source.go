package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ datasource.DataSource = &VpnConnectionDataSource{}

type VpnConnectionDataSource struct {
	client *AutoLoginClient
}

type VpnConnectionDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Status         types.String `tfsdk:"status"`
	LocalIP        types.String `tfsdk:"local_ip"`
	RemoteIP       types.String `tfsdk:"remote_ip"`
	RemoteNetwork  types.String `tfsdk:"remote_network"`
	BytesIn        types.Int64  `tfsdk:"bytes_in"`
	BytesOut       types.Int64  `tfsdk:"bytes_out"`
	ConnectedSince types.Int64  `tfsdk:"connected_since"`
}

func NewVpnConnectionDataSource() datasource.DataSource {
	return &VpnConnectionDataSource{}
}

func (d *VpnConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_connection"
}

func (d *VpnConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a VPN connection. Lookup by ID or name. Read-only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the VPN connection. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the VPN connection. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"type":            schema.StringAttribute{Computed: true},
			"status":          schema.StringAttribute{Computed: true},
			"local_ip":        schema.StringAttribute{Computed: true},
			"remote_ip":       schema.StringAttribute{Computed: true},
			"remote_network":  schema.StringAttribute{Computed: true},
			"bytes_in":        schema.Int64Attribute{Computed: true},
			"bytes_out":       schema.Int64Attribute{Computed: true},
			"connected_since": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *VpnConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpnConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VpnConnectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either 'id' or 'name' must be specified.")
		return
	}

	connections, err := d.client.ListVpnConnections(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "VPN connections")
		return
	}

	var found *unifi.VpnConnection
	if hasID {
		id := config.ID.ValueString()
		for i := range connections {
			if connections[i].ID == id {
				found = &connections[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("VPN Connection Not Found", fmt.Sprintf("No VPN connection found with ID '%s'.", id))
			return
		}
	} else {
		name := config.Name.ValueString()
		for i := range connections {
			if connections[i].Name == name {
				found = &connections[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("VPN Connection Not Found", fmt.Sprintf("No VPN connection found with name '%s'.", name))
			return
		}
	}

	config.ID = types.StringValue(found.ID)
	config.Name = stringValueOrNull(found.Name)
	config.Type = stringValueOrNull(found.Type)
	config.Status = stringValueOrNull(found.Status)
	config.LocalIP = stringValueOrNull(found.LocalIP)
	config.RemoteIP = stringValueOrNull(found.RemoteIP)
	config.RemoteNetwork = stringValueOrNull(found.RemoteNetwork)

	if found.BytesIn != nil {
		config.BytesIn = types.Int64Value(*found.BytesIn)
	} else {
		config.BytesIn = types.Int64Null()
	}
	if found.BytesOut != nil {
		config.BytesOut = types.Int64Value(*found.BytesOut)
	} else {
		config.BytesOut = types.Int64Null()
	}
	if found.ConnectedSince != nil {
		config.ConnectedSince = types.Int64Value(*found.ConnectedSince)
	} else {
		config.ConnectedSince = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
