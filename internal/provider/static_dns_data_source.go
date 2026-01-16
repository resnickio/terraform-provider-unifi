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

var _ datasource.DataSource = &StaticDNSDataSource{}

type StaticDNSDataSource struct {
	client *AutoLoginClient
}

type StaticDNSDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Key        types.String `tfsdk:"key"`
	Value      types.String `tfsdk:"value"`
	RecordType types.String `tfsdk:"record_type"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	TTL        types.Int64  `tfsdk:"ttl"`
	Port       types.Int64  `tfsdk:"port"`
	Priority   types.Int64  `tfsdk:"priority"`
	Weight     types.Int64  `tfsdk:"weight"`
}

func NewStaticDNSDataSource() datasource.DataSource {
	return &StaticDNSDataSource{}
}

func (d *StaticDNSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_dns"
}

func (d *StaticDNSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi static DNS record. Lookup by either id or key (hostname).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the static DNS record. Specify either id or key.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("key")),
				},
			},
			"key": schema.StringAttribute{
				Description: "The hostname or domain name for the DNS record. Specify either id or key.",
				Optional:    true,
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value for the DNS record (IP address, hostname, or other value depending on record type).",
				Computed:    true,
			},
			"record_type": schema.StringAttribute{
				Description: "The DNS record type (A, AAAA, CNAME, MX, NS, TXT, SRV).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the DNS record is enabled.",
				Computed:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live in seconds for the DNS record.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port number for SRV records.",
				Computed:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "Priority value for MX and SRV records.",
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Weight value for SRV records.",
				Computed:    true,
			},
		},
	}
}

func (d *StaticDNSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StaticDNSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StaticDNSDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasKey := !config.Key.IsNull() && config.Key.ValueString() != ""

	if !hasID && !hasKey {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'key' must be specified to look up a static DNS record.",
		)
		return
	}

	var dns *unifi.StaticDNS
	var err error

	if hasID {
		dns, err = d.client.GetStaticDNS(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "static DNS record")
			return
		}
	} else {
		records, err := d.client.ListStaticDNS(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "static DNS records")
			return
		}

		searchKey := config.Key.ValueString()
		for i := range records {
			if records[i].Key == searchKey {
				dns = &records[i]
				break
			}
		}

		if dns == nil {
			resp.Diagnostics.AddError(
				"Static DNS Record Not Found",
				fmt.Sprintf("No static DNS record found with key '%s'.", searchKey),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, dns, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *StaticDNSDataSource) sdkToState(ctx context.Context, dns *unifi.StaticDNS, state *StaticDNSDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(dns.ID)
	state.Key = types.StringValue(dns.Key)
	state.Value = types.StringValue(dns.Value)
	state.RecordType = types.StringValue(dns.RecordType)
	state.Enabled = types.BoolValue(derefBool(dns.Enabled))

	if dns.TTL != nil {
		state.TTL = types.Int64Value(int64(*dns.TTL))
	} else {
		state.TTL = types.Int64Null()
	}

	if dns.Port != nil {
		state.Port = types.Int64Value(int64(*dns.Port))
	} else {
		state.Port = types.Int64Null()
	}

	if dns.Priority != nil {
		state.Priority = types.Int64Value(int64(*dns.Priority))
	} else {
		state.Priority = types.Int64Null()
	}

	if dns.Weight != nil {
		state.Weight = types.Int64Value(int64(*dns.Weight))
	} else {
		state.Weight = types.Int64Null()
	}

	return diags
}
