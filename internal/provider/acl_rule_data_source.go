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

var _ datasource.DataSource = &AclRuleDataSource{}

type AclRuleDataSource struct {
	client *AutoLoginClient
}

type AclRuleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
}

func NewAclRuleDataSource() datasource.DataSource {
	return &AclRuleDataSource{}
}

func (d *AclRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_rule"
}

func (d *AclRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing ACL rule. Lookup by ID or name. Read-only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the ACL rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the ACL rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"enabled":     schema.BoolAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *AclRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AclRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AclRuleDataSourceModel
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

	rules, err := d.client.ListAclRules(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "ACL rules")
		return
	}

	var found *unifi.AclRule
	if hasID {
		id := config.ID.ValueString()
		for i := range rules {
			if rules[i].ID == id {
				found = &rules[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("ACL Rule Not Found", fmt.Sprintf("No ACL rule found with ID '%s'.", id))
			return
		}
	} else {
		name := config.Name.ValueString()
		for i := range rules {
			if rules[i].Name == name {
				found = &rules[i]
				break
			}
		}
		if found == nil {
			resp.Diagnostics.AddError("ACL Rule Not Found", fmt.Sprintf("No ACL rule found with name '%s'.", name))
			return
		}
	}

	config.ID = types.StringValue(found.ID)
	config.Name = stringValueOrNull(found.Name)
	config.Enabled = types.BoolValue(derefBool(found.Enabled))
	config.Description = stringValueOrNull(found.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
