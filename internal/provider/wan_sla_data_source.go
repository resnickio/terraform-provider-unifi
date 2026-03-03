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

var _ datasource.DataSource = &WanSlaDataSource{}

type WanSlaDataSource struct {
	client *AutoLoginClient
}

type WanSlaDataSourceModel struct {
	ID                  types.String  `tfsdk:"id"`
	Name                types.String  `tfsdk:"name"`
	Enabled             types.Bool    `tfsdk:"enabled"`
	Interface           types.String  `tfsdk:"interface"`
	Target              types.String  `tfsdk:"target"`
	ThresholdLatency    types.Int64   `tfsdk:"threshold_latency"`
	ThresholdPacketLoss types.Float64 `tfsdk:"threshold_packet_loss"`
}

func NewWanSlaDataSource() datasource.DataSource {
	return &WanSlaDataSource{}
}

func (d *WanSlaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan_sla"
}

func (d *WanSlaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a WAN SLA monitor. Lookup by ID or name. Read-only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the WAN SLA. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the WAN SLA. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"enabled":               schema.BoolAttribute{Computed: true},
			"interface":             schema.StringAttribute{Computed: true},
			"target":                schema.StringAttribute{Computed: true},
			"threshold_latency":     schema.Int64Attribute{Computed: true},
			"threshold_packet_loss": schema.Float64Attribute{Computed: true},
		},
	}
}

func (d *WanSlaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WanSlaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WanSlaDataSourceModel
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

	slas, err := d.client.ListWanSlas(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "WAN SLAs")
		return
	}

	var found *unifi.WanSla
	if hasID {
		id := config.ID.ValueString()
		for i := range slas {
			if slas[i].ID == id {
				found = &slas[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("WAN SLA Not Found", fmt.Sprintf("No WAN SLA found with ID '%s'.", id))
			return
		}
	} else {
		name := config.Name.ValueString()
		for i := range slas {
			if slas[i].Name == name {
				found = &slas[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("WAN SLA Not Found", fmt.Sprintf("No WAN SLA found with name '%s'.", name))
			return
		}
	}

	config.ID = types.StringValue(found.ID)
	config.Name = stringValueOrNull(found.Name)
	config.Enabled = types.BoolValue(derefBool(found.Enabled))
	config.Interface = stringValueOrNull(found.Interface)
	config.Target = stringValueOrNull(found.Target)

	if found.ThresholdLatency != nil {
		config.ThresholdLatency = types.Int64Value(int64(*found.ThresholdLatency))
	} else {
		config.ThresholdLatency = types.Int64Null()
	}
	if found.ThresholdPacketLoss != nil {
		config.ThresholdPacketLoss = types.Float64Value(*found.ThresholdPacketLoss)
	} else {
		config.ThresholdPacketLoss = types.Float64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
