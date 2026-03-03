package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ resource.Resource = &ContentFilteringResource{}

type ContentFilteringResource struct {
	client *AutoLoginClient
}

type ContentFilteringResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Enabled           types.Bool     `tfsdk:"enabled"`
	BlockedCategories types.Set      `tfsdk:"blocked_categories"`
	AllowedDomains    types.Set      `tfsdk:"allowed_domains"`
	BlockedDomains    types.Set      `tfsdk:"blocked_domains"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func NewContentFilteringResource() resource.Resource {
	return &ContentFilteringResource{}
}

func (r *ContentFilteringResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_filtering"
}

func (r *ContentFilteringResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi content filtering configuration. " +
			"This is a singleton resource — one per site. Delete resets to defaults (disabled).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable content filtering.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"blocked_categories": schema.SetAttribute{
				Description: "Set of blocked content categories.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"allowed_domains": schema.SetAttribute{
				Description: "Set of allowed domains (bypass filtering).",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"blocked_domains": schema.SetAttribute{
				Description: "Set of explicitly blocked domains.",
				Optional:    true,
				Computed:    true,
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

func (r *ContentFilteringResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *ContentFilteringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContentFilteringResourceModel
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

	config := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateContentFiltering(ctx, config)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "content filtering")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContentFilteringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContentFilteringResourceModel
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

	filtering, err := r.client.GetContentFiltering(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "content filtering")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, filtering, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContentFilteringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContentFilteringResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	config := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateContentFiltering(ctx, config)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "content filtering")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContentFilteringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContentFilteringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	defaults := &unifi.ContentFiltering{
		Enabled:           boolPtr(false),
		BlockedCategories: []string{},
		AllowedDomains:    []string{},
		BlockedDomains:    []string{},
	}

	_, err := r.client.UpdateContentFiltering(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "content filtering")
		return
	}
}

func (r *ContentFilteringResource) planToSDK(ctx context.Context, plan *ContentFilteringResourceModel, diags *diag.Diagnostics) *unifi.ContentFiltering {
	config := &unifi.ContentFiltering{
		Enabled: boolPtr(plan.Enabled.ValueBool()),
	}

	if !plan.BlockedCategories.IsNull() && !plan.BlockedCategories.IsUnknown() {
		var categories []string
		diags.Append(plan.BlockedCategories.ElementsAs(ctx, &categories, false)...)
		config.BlockedCategories = categories
	}
	if !plan.AllowedDomains.IsNull() && !plan.AllowedDomains.IsUnknown() {
		var domains []string
		diags.Append(plan.AllowedDomains.ElementsAs(ctx, &domains, false)...)
		config.AllowedDomains = domains
	}
	if !plan.BlockedDomains.IsNull() && !plan.BlockedDomains.IsUnknown() {
		var domains []string
		diags.Append(plan.BlockedDomains.ElementsAs(ctx, &domains, false)...)
		config.BlockedDomains = domains
	}

	return config
}

func (r *ContentFilteringResource) sdkToState(ctx context.Context, filtering *unifi.ContentFiltering, state *ContentFilteringResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue("content_filtering")
	state.Enabled = types.BoolValue(derefBool(filtering.Enabled))

	if len(filtering.BlockedCategories) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, filtering.BlockedCategories)
		diags.Append(d...)
		state.BlockedCategories = s
	} else {
		state.BlockedCategories = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if len(filtering.AllowedDomains) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, filtering.AllowedDomains)
		diags.Append(d...)
		state.AllowedDomains = s
	} else {
		state.AllowedDomains = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if len(filtering.BlockedDomains) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, filtering.BlockedDomains)
		diags.Append(d...)
		state.BlockedDomains = s
	} else {
		state.BlockedDomains = types.SetValueMust(types.StringType, []attr.Value{})
	}

	return diags
}
