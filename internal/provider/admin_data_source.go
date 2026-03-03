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

var _ datasource.DataSource = &AdminDataSource{}

type AdminDataSource struct {
	client *AutoLoginClient
}

type AdminDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Email        types.String `tfsdk:"email"`
	Role         types.String `tfsdk:"role"`
	IsSuperAdmin types.Bool   `tfsdk:"is_super_admin"`
	IsVerified   types.Bool   `tfsdk:"is_verified"`
	DeviceID     types.String `tfsdk:"device_id"`
	TimeCreated  types.Int64  `tfsdk:"time_created"`
}

func NewAdminDataSource() datasource.DataSource {
	return &AdminDataSource{}
}

func (d *AdminDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin"
}

func (d *AdminDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a UniFi controller admin by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the admin. Use `id` or `name` to look up the admin.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the admin.",
				Optional:    true,
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the admin.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role of the admin.",
				Computed:    true,
			},
			"is_super_admin": schema.BoolAttribute{
				Description: "Whether this admin is a super admin.",
				Computed:    true,
			},
			"is_verified": schema.BoolAttribute{
				Description: "Whether this admin is verified.",
				Computed:    true,
			},
			"device_id": schema.StringAttribute{
				Description: "The device ID associated with this admin.",
				Computed:    true,
			},
			"time_created": schema.Int64Attribute{
				Description: "Unix timestamp when the admin was created.",
				Computed:    true,
			},
		},
	}
}

func (d *AdminDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AdminDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AdminDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	admins, err := d.client.ListAdmins(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "admins")
		return
	}

	var found *unifi.Admin

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		id := config.ID.ValueString()
		for i := range admins {
			if admins[i].ID == id {
				found = &admins[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("Admin not found", fmt.Sprintf("No admin found with ID %q.", id))
			return
		}
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		for i := range admins {
			if admins[i].Name == name {
				found = &admins[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("Admin not found", fmt.Sprintf("No admin found with name %q.", name))
			return
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Admin not found", "Either id or name must be specified.")
		return
	}

	var state AdminDataSourceModel
	state.ID = types.StringValue(found.ID)
	state.Name = stringValueOrNull(found.Name)
	state.Email = stringValueOrNull(found.Email)
	state.Role = stringValueOrNull(found.Role)
	state.IsSuperAdmin = types.BoolValue(derefBool(found.IsSuperAdmin))
	state.IsVerified = types.BoolValue(derefBool(found.IsVerified))
	state.DeviceID = stringValueOrNull(found.DeviceID)

	if found.TimeCreated != nil {
		state.TimeCreated = types.Int64Value(*found.TimeCreated)
	} else {
		state.TimeCreated = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
