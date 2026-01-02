package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &FirewallGroupResource{}
	_ resource.ResourceWithImportState = &FirewallGroupResource{}
)

type FirewallGroupResource struct {
	client *AutoLoginClient
}

type FirewallGroupResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	SiteID    types.String   `tfsdk:"site_id"`
	Name      types.String   `tfsdk:"name"`
	GroupType types.String   `tfsdk:"group_type"`
	Members   types.Set      `tfsdk:"members"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func NewFirewallGroupResource() resource.Resource {
	return &FirewallGroupResource{}
}

func (r *FirewallGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_group"
}

func (r *FirewallGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi firewall group (IP address group, port group, or IPv6 address group).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the firewall group is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall group.",
				Required:    true,
			},
			"group_type": schema.StringAttribute{
				Description: "The type of the firewall group. Valid values are: " +
					"'address-group' (IPv4 addresses/CIDRs), " +
					"'port-group' (port numbers/ranges), " +
					"'ipv6-address-group' (IPv6 addresses/CIDRs).",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("address-group", "port-group", "ipv6-address-group"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"members": schema.SetAttribute{
				Description: "The members of the firewall group. For address groups, this is a set of " +
					"IP addresses or CIDR ranges. For port groups, this is a set of port numbers or ranges (e.g., '80', '8080-8090').",
				Required:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *FirewallGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *FirewallGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	var members []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := &unifi.FirewallGroup{
		Name:         plan.Name.ValueString(),
		GroupType:    plan.GroupType.ValueString(),
		GroupMembers: members,
	}

	created, err := r.client.CreateFirewallGroup(ctx, group)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "firewall group")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, 2*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	group, err := r.client.GetFirewallGroup(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "firewall group")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, group, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FirewallGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FirewallGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var members []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := &unifi.FirewallGroup{
		ID:           state.ID.ValueString(),
		SiteID:       state.SiteID.ValueString(),
		Name:         plan.Name.ValueString(),
		GroupType:    plan.GroupType.ValueString(),
		GroupMembers: members,
	}

	updated, err := r.client.UpdateFirewallGroup(ctx, state.ID.ValueString(), group)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "firewall group")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	err := r.client.DeleteFirewallGroup(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "firewall group")
		return
	}
}

func (r *FirewallGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// sdkToState updates the Terraform state from an SDK FirewallGroup struct.
func (r *FirewallGroupResource) sdkToState(ctx context.Context, group *unifi.FirewallGroup, state *FirewallGroupResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(group.ID)
	state.SiteID = types.StringValue(group.SiteID)
	state.Name = types.StringValue(group.Name)
	state.GroupType = types.StringValue(group.GroupType)

	members, d := types.SetValueFrom(ctx, types.StringType, group.GroupMembers)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	state.Members = members

	return diags
}
