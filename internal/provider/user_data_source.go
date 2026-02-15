package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ datasource.DataSource = &UserDataSource{}

type UserDataSource struct {
	client *AutoLoginClient
}

type UserDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SiteID                types.String `tfsdk:"site_id"`
	MAC                   types.String `tfsdk:"mac"`
	Name                  types.String `tfsdk:"name"`
	Note                  types.String `tfsdk:"note"`
	Noted                 types.Bool   `tfsdk:"noted"`
	UseFixedIP            types.Bool   `tfsdk:"use_fixed_ip"`
	FixedIP               types.String `tfsdk:"fixed_ip"`
	NetworkID             types.String `tfsdk:"network_id"`
	LocalDnsRecord        types.String `tfsdk:"local_dns_record"`
	LocalDnsRecordEnabled types.Bool   `tfsdk:"local_dns_record_enabled"`
	UsergroupID           types.String `tfsdk:"usergroup_id"`
	Blocked               types.Bool   `tfsdk:"blocked"`
	IP                    types.String `tfsdk:"ip"`
	Hostname              types.String `tfsdk:"hostname"`
	OUI                   types.String `tfsdk:"oui"`
	FirstSeen             types.Int64  `tfsdk:"first_seen"`
	LastSeen              types.Int64  `tfsdk:"last_seen"`
}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi user (client device record). Lookup by either id or mac.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user. Specify either id or mac.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("mac")),
				},
			},
			"mac": schema.StringAttribute{
				Description: "The MAC address of the client device (format: aa:bb:cc:dd:ee:ff). Specify either id or mac.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the user exists.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "A friendly name for the client device.",
				Computed:    true,
			},
			"note": schema.StringAttribute{
				Description: "Notes or description for the client device.",
				Computed:    true,
			},
			"noted": schema.BoolAttribute{
				Description: "Whether the device has a note.",
				Computed:    true,
			},
			"use_fixed_ip": schema.BoolAttribute{
				Description: "Whether DHCP reservation is enabled for this device.",
				Computed:    true,
			},
			"fixed_ip": schema.StringAttribute{
				Description: "The fixed IP address for DHCP reservation.",
				Computed:    true,
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID for the fixed IP DHCP reservation.",
				Computed:    true,
			},
			"local_dns_record": schema.StringAttribute{
				Description: "A local DNS hostname record for this device.",
				Computed:    true,
			},
			"local_dns_record_enabled": schema.BoolAttribute{
				Description: "Whether the local DNS record is enabled.",
				Computed:    true,
			},
			"usergroup_id": schema.StringAttribute{
				Description: "The user group ID for bandwidth limiting.",
				Computed:    true,
			},
			"blocked": schema.BoolAttribute{
				Description: "Whether the device is blocked from network access.",
				Computed:    true,
			},
			"ip": schema.StringAttribute{
				Description: "The current IP address of the device.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the device from DHCP.",
				Computed:    true,
			},
			"oui": schema.StringAttribute{
				Description: "The manufacturer OUI code derived from the MAC address.",
				Computed:    true,
			},
			"first_seen": schema.Int64Attribute{
				Description: "Unix timestamp when the device was first seen.",
				Computed:    true,
			},
			"last_seen": schema.Int64Attribute{
				Description: "Unix timestamp when the device was last seen.",
				Computed:    true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasMAC := !config.MAC.IsNull() && config.MAC.ValueString() != ""

	if !hasID && !hasMAC {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'mac' must be specified to look up a user.",
		)
		return
	}

	var user *unifi.User
	var err error

	if hasID {
		user, err = d.client.GetUser(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "user")
			return
		}
	} else {
		users, err := d.client.ListUsers(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "users")
			return
		}

		searchMAC := strings.ToLower(config.MAC.ValueString())
		for i := range users {
			if strings.ToLower(users[i].MAC) == searchMAC {
				user = &users[i]
				break
			}
		}

		if user == nil {
			resp.Diagnostics.AddError(
				"User Not Found",
				fmt.Sprintf("No user found with MAC address '%s'.", config.MAC.ValueString()),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(user, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *UserDataSource) sdkToState(user *unifi.User, state *UserDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(user.ID)
	state.SiteID = types.StringValue(user.SiteID)
	state.MAC = types.StringValue(user.MAC)
	state.Name = stringValueOrNull(user.Name)
	state.Note = stringValueOrNull(user.Note)

	if user.Noted != nil {
		state.Noted = types.BoolValue(*user.Noted)
	} else {
		state.Noted = types.BoolNull()
	}

	if user.UseFixedIP != nil {
		state.UseFixedIP = types.BoolValue(*user.UseFixedIP)
	} else {
		state.UseFixedIP = types.BoolNull()
	}

	state.FixedIP = stringValueOrNull(user.FixedIP)
	state.NetworkID = stringValueOrNull(user.NetworkID)
	state.LocalDnsRecord = stringValueOrNull(user.LocalDnsRecord)

	if user.LocalDnsRecordEnabled != nil {
		state.LocalDnsRecordEnabled = types.BoolValue(*user.LocalDnsRecordEnabled)
	} else {
		state.LocalDnsRecordEnabled = types.BoolNull()
	}

	state.UsergroupID = stringValueOrNull(user.UsergroupID)

	if user.Blocked != nil {
		state.Blocked = types.BoolValue(*user.Blocked)
	} else {
		state.Blocked = types.BoolNull()
	}

	state.IP = stringValueOrNull(user.IP)
	state.Hostname = stringValueOrNull(user.Hostname)
	state.OUI = stringValueOrNull(user.OUI)

	if user.FirstSeen != nil {
		state.FirstSeen = types.Int64Value(*user.FirstSeen)
	} else {
		state.FirstSeen = types.Int64Null()
	}

	if user.LastSeen != nil {
		state.LastSeen = types.Int64Value(*user.LastSeen)
	} else {
		state.LastSeen = types.Int64Null()
	}

	return diags
}
