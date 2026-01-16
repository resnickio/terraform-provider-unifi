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

var _ datasource.DataSource = &PortForwardDataSource{}

type PortForwardDataSource struct {
	client *AutoLoginClient
}

type PortForwardDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	SiteID        types.String `tfsdk:"site_id"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Protocol      types.String `tfsdk:"protocol"`
	DstPort       types.String `tfsdk:"dst_port"`
	FwdPort       types.String `tfsdk:"fwd_port"`
	FwdIP         types.String `tfsdk:"fwd_ip"`
	Src           types.String `tfsdk:"src"`
	PfwdInterface types.String `tfsdk:"pfwd_interface"`
	Log           types.Bool   `tfsdk:"log"`
}

func NewPortForwardDataSource() datasource.DataSource {
	return &PortForwardDataSource{}
}

func (d *PortForwardDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_forward"
}

func (d *PortForwardDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi port forward rule. Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the port forward rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the port forward rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the port forward rule exists.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the port forward rule is enabled.",
				Computed:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol for port forwarding (tcp, udp, tcp_udp).",
				Computed:    true,
			},
			"dst_port": schema.StringAttribute{
				Description: "The destination port to forward from (external port).",
				Computed:    true,
			},
			"fwd_port": schema.StringAttribute{
				Description: "The port to forward to on the destination host.",
				Computed:    true,
			},
			"fwd_ip": schema.StringAttribute{
				Description: "The IP address to forward traffic to.",
				Computed:    true,
			},
			"src": schema.StringAttribute{
				Description: "Source IP/CIDR restriction for the port forward.",
				Computed:    true,
			},
			"pfwd_interface": schema.StringAttribute{
				Description: "The WAN interface for the port forward (wan, wan2, both).",
				Computed:    true,
			},
			"log": schema.BoolAttribute{
				Description: "Whether to log forwarded traffic.",
				Computed:    true,
			},
		},
	}
}

func (d *PortForwardDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PortForwardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PortForwardDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a port forward rule.",
		)
		return
	}

	var pf *unifi.PortForward
	var err error

	if hasID {
		pf, err = d.client.GetPortForward(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "port forward")
			return
		}
	} else {
		forwards, err := d.client.ListPortForwards(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "port forwards")
			return
		}

		searchName := config.Name.ValueString()
		for i := range forwards {
			if forwards[i].Name == searchName {
				pf = &forwards[i]
				break
			}
		}

		if pf == nil {
			resp.Diagnostics.AddError(
				"Port Forward Not Found",
				fmt.Sprintf("No port forward rule found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, pf, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *PortForwardDataSource) sdkToState(ctx context.Context, pf *unifi.PortForward, state *PortForwardDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(pf.ID)
	state.SiteID = types.StringValue(pf.SiteID)
	state.Name = types.StringValue(pf.Name)
	state.Enabled = types.BoolValue(derefBool(pf.Enabled))
	state.Protocol = types.StringValue(pf.Proto)
	state.DstPort = types.StringValue(pf.DstPort)
	state.FwdPort = types.StringValue(pf.FwdPort)
	state.FwdIP = types.StringValue(pf.Fwd)
	state.Src = stringValueOrNull(pf.Src)

	if pf.PfwdInterface != "" {
		state.PfwdInterface = types.StringValue(pf.PfwdInterface)
	} else {
		state.PfwdInterface = types.StringValue("wan")
	}

	state.Log = types.BoolValue(derefBool(pf.Log))

	return diags
}
