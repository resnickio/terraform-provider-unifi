package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &UserGroupResource{}
	_ resource.ResourceWithImportState = &UserGroupResource{}
)

type UserGroupResource struct {
	client *AutoLoginClient
}

type UserGroupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	SiteID         types.String `tfsdk:"site_id"`
	Name           types.String `tfsdk:"name"`
	QosRateMaxDown types.Int64  `tfsdk:"qos_rate_max_down"`
	QosRateMaxUp   types.Int64  `tfsdk:"qos_rate_max_up"`
}

func NewUserGroupResource() resource.Resource {
	return &UserGroupResource{}
}

func (r *UserGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (r *UserGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi user group (bandwidth profile). User groups define bandwidth limits for clients.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the user group is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the user group.",
				Required:    true,
			},
			"qos_rate_max_down": schema.Int64Attribute{
				Description: "Maximum download rate in kbps. Set to -1 for unlimited.",
				Optional:    true,
			},
			"qos_rate_max_up": schema.Int64Attribute{
				Description: "Maximum upload rate in kbps. Set to -1 for unlimited.",
				Optional:    true,
			},
		},
	}
}

func (r *UserGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := r.planToSDK(&plan)

	created, err := r.client.CreateUserGroup(ctx, group)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "user group")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetUserGroup(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "user group")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(group, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := r.planToSDK(&plan)
	group.ID = state.ID.ValueString()
	group.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateUserGroup(ctx, state.ID.ValueString(), group)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "user group")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUserGroup(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "user group")
		return
	}
}

func (r *UserGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UserGroupResource) planToSDK(plan *UserGroupResourceModel) *unifi.UserGroup {
	group := &unifi.UserGroup{
		Name: plan.Name.ValueString(),
	}

	if !plan.QosRateMaxDown.IsNull() && !plan.QosRateMaxDown.IsUnknown() {
		group.QosRateMaxDown = intPtr(plan.QosRateMaxDown.ValueInt64())
	}

	if !plan.QosRateMaxUp.IsNull() && !plan.QosRateMaxUp.IsUnknown() {
		group.QosRateMaxUp = intPtr(plan.QosRateMaxUp.ValueInt64())
	}

	return group
}

func (r *UserGroupResource) sdkToState(group *unifi.UserGroup, state *UserGroupResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(group.ID)
	state.SiteID = types.StringValue(group.SiteID)
	state.Name = types.StringValue(group.Name)

	if group.QosRateMaxDown != nil {
		state.QosRateMaxDown = types.Int64Value(int64(*group.QosRateMaxDown))
	} else {
		state.QosRateMaxDown = types.Int64Null()
	}

	if group.QosRateMaxUp != nil {
		state.QosRateMaxUp = types.Int64Value(int64(*group.QosRateMaxUp))
	} else {
		state.QosRateMaxUp = types.Int64Null()
	}

	return diags
}
