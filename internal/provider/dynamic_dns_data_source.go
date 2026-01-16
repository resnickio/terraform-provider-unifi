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

var _ datasource.DataSource = &DynamicDNSDataSource{}

type DynamicDNSDataSource struct {
	client *AutoLoginClient
}

type DynamicDNSDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	SiteID    types.String `tfsdk:"site_id"`
	Service   types.String `tfsdk:"service"`
	HostName  types.String `tfsdk:"hostname"`
	Login     types.String `tfsdk:"login"`
	Server    types.String `tfsdk:"server"`
	Interface types.String `tfsdk:"interface"`
	Options   types.String `tfsdk:"options"`
}

func NewDynamicDNSDataSource() datasource.DataSource {
	return &DynamicDNSDataSource{}
}

func (d *DynamicDNSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_dns"
}

func (d *DynamicDNSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi dynamic DNS configuration. Lookup by either id or hostname.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the dynamic DNS configuration. Specify either id or hostname.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("hostname")),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname to update with the dynamic DNS service. Specify either id or hostname.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the dynamic DNS is configured.",
				Computed:    true,
			},
			"service": schema.StringAttribute{
				Description: "The dynamic DNS service provider.",
				Computed:    true,
			},
			"login": schema.StringAttribute{
				Description: "The login/username for the dynamic DNS service.",
				Computed:    true,
			},
			"server": schema.StringAttribute{
				Description: "The server address for the dynamic DNS service.",
				Computed:    true,
			},
			"interface": schema.StringAttribute{
				Description: "The WAN interface to monitor for IP changes.",
				Computed:    true,
			},
			"options": schema.StringAttribute{
				Description: "Additional options for the dynamic DNS service.",
				Computed:    true,
			},
		},
	}
}

func (d *DynamicDNSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DynamicDNSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DynamicDNSDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasHostname := !config.HostName.IsNull() && config.HostName.ValueString() != ""

	if !hasID && !hasHostname {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'hostname' must be specified to look up a dynamic DNS configuration.",
		)
		return
	}

	var dns *unifi.DynamicDNS
	var err error

	if hasID {
		dns, err = d.client.GetDynamicDNS(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "dynamic DNS configuration")
			return
		}
	} else {
		records, err := d.client.ListDynamicDNS(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "dynamic DNS configurations")
			return
		}

		searchHostname := config.HostName.ValueString()
		for i := range records {
			if records[i].HostName == searchHostname {
				dns = &records[i]
				break
			}
		}

		if dns == nil {
			resp.Diagnostics.AddError(
				"Dynamic DNS Configuration Not Found",
				fmt.Sprintf("No dynamic DNS configuration found with hostname '%s'.", searchHostname),
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

func (d *DynamicDNSDataSource) sdkToState(ctx context.Context, dns *unifi.DynamicDNS, state *DynamicDNSDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(dns.ID)
	state.SiteID = stringValueOrNull(dns.SiteID)
	state.Service = types.StringValue(dns.Service)
	state.HostName = types.StringValue(dns.HostName)
	state.Interface = stringValueOrNull(dns.Interface)

	if dns.Login != "" {
		state.Login = types.StringValue(dns.Login)
	} else {
		state.Login = types.StringNull()
	}

	if dns.Server != "" {
		state.Server = types.StringValue(dns.Server)
	} else {
		state.Server = types.StringNull()
	}

	if dns.Options != "" {
		state.Options = types.StringValue(dns.Options)
	} else {
		state.Options = types.StringNull()
	}

	return diags
}
