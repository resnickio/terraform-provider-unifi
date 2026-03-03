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

var _ datasource.DataSource = &ActiveClientDataSource{}

type ActiveClientDataSource struct {
	client *AutoLoginClient
}

type ActiveClientDataSourceModel struct {
	ID           types.String  `tfsdk:"id"`
	MAC          types.String  `tfsdk:"mac"`
	DisplayName  types.String  `tfsdk:"display_name"`
	Status       types.String  `tfsdk:"status"`
	Type         types.String  `tfsdk:"type"`
	IsWired      types.Bool    `tfsdk:"is_wired"`
	IsGuest      types.Bool    `tfsdk:"is_guest"`
	Blocked      types.Bool    `tfsdk:"blocked"`
	NetworkID    types.String  `tfsdk:"network_id"`
	NetworkName  types.String  `tfsdk:"network_name"`
	LastIP       types.String  `tfsdk:"last_ip"`
	VLAN         types.Int64   `tfsdk:"vlan"`
	OUI          types.String  `tfsdk:"oui"`
	Uptime       types.Int64   `tfsdk:"uptime"`
	RxBytes      types.Int64   `tfsdk:"rx_bytes"`
	TxBytes      types.Int64   `tfsdk:"tx_bytes"`
	Satisfaction types.Float64 `tfsdk:"satisfaction"`
	SiteID       types.String  `tfsdk:"site_id"`
}

func NewActiveClientDataSource() datasource.DataSource {
	return &ActiveClientDataSource{}
}

func (d *ActiveClientDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_active_client"
}

func (d *ActiveClientDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a currently active (connected) client. Lookup by MAC address or display name.",
		Attributes: map[string]schema.Attribute{
			"mac": schema.StringAttribute{
				Description: "The MAC address of the client. Specify either mac or display_name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("display_name")),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the client. Specify either mac or display_name.",
				Optional:    true,
				Computed:    true,
			},
			"id":           schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"type":         schema.StringAttribute{Computed: true},
			"is_wired":     schema.BoolAttribute{Computed: true},
			"is_guest":     schema.BoolAttribute{Computed: true},
			"blocked":      schema.BoolAttribute{Computed: true},
			"network_id":   schema.StringAttribute{Computed: true},
			"network_name": schema.StringAttribute{Computed: true},
			"last_ip":      schema.StringAttribute{Computed: true},
			"vlan":         schema.Int64Attribute{Computed: true},
			"oui":          schema.StringAttribute{Computed: true},
			"uptime":       schema.Int64Attribute{Computed: true},
			"rx_bytes":     schema.Int64Attribute{Computed: true},
			"tx_bytes":     schema.Int64Attribute{Computed: true},
			"satisfaction": schema.Float64Attribute{Computed: true},
			"site_id":      schema.StringAttribute{Computed: true},
		},
	}
}

func (d *ActiveClientDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ActiveClientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ActiveClientDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasMAC := !config.MAC.IsNull() && config.MAC.ValueString() != ""
	hasName := !config.DisplayName.IsNull() && config.DisplayName.ValueString() != ""

	if !hasMAC && !hasName {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either 'mac' or 'display_name' must be specified.")
		return
	}

	clients, err := d.client.ListActiveClients(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "active clients")
		return
	}

	var found *unifi.Client
	if hasMAC {
		mac := config.MAC.ValueString()
		for i := range clients {
			if clients[i].MAC == mac {
				found = &clients[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("Active Client Not Found", fmt.Sprintf("No active client found with MAC '%s'.", mac))
			return
		}
	} else {
		name := config.DisplayName.ValueString()
		for i := range clients {
			if clients[i].DisplayName == name {
				found = &clients[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("Active Client Not Found", fmt.Sprintf("No active client found with display name '%s'.", name))
			return
		}
	}

	d.sdkToState(found, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *ActiveClientDataSource) sdkToState(client *unifi.Client, state *ActiveClientDataSourceModel) {
	state.ID = types.StringValue(client.ID)
	state.MAC = types.StringValue(client.MAC)
	state.DisplayName = stringValueOrNull(client.DisplayName)
	state.Status = stringValueOrNull(client.Status)
	state.Type = stringValueOrNull(client.Type)
	state.IsWired = types.BoolValue(derefBool(client.IsWired))
	state.IsGuest = types.BoolValue(derefBool(client.IsGuest))
	state.Blocked = types.BoolValue(derefBool(client.Blocked))
	state.NetworkID = stringValueOrNull(client.NetworkID)
	state.NetworkName = stringValueOrNull(client.NetworkName)
	state.LastIP = stringValueOrNull(client.LastIP)
	state.OUI = stringValueOrNull(client.OUI)
	state.SiteID = stringValueOrNull(client.SiteID)

	if client.VLAN != nil {
		state.VLAN = types.Int64Value(int64(*client.VLAN))
	} else {
		state.VLAN = types.Int64Null()
	}
	if client.Uptime != nil {
		state.Uptime = types.Int64Value(*client.Uptime)
	} else {
		state.Uptime = types.Int64Null()
	}
	if client.RxBytes != nil {
		state.RxBytes = types.Int64Value(*client.RxBytes)
	} else {
		state.RxBytes = types.Int64Null()
	}
	if client.TxBytes != nil {
		state.TxBytes = types.Int64Value(*client.TxBytes)
	} else {
		state.TxBytes = types.Int64Null()
	}
	if client.Satisfaction != nil {
		state.Satisfaction = types.Float64Value(*client.Satisfaction)
	} else {
		state.Satisfaction = types.Float64Null()
	}
}
