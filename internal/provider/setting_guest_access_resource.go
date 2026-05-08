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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &SettingGuestAccessResource{}
	_ resource.ResourceWithImportState = &SettingGuestAccessResource{}
)

type SettingGuestAccessResource struct {
	client *AutoLoginClient
}

type SettingGuestAccessResourceModel struct {
	ID                                 types.String   `tfsdk:"id"`
	SiteID                             types.String   `tfsdk:"site_id"`
	PortalEnabled                      types.Bool     `tfsdk:"portal_enabled"`
	PortalCustomized                   types.Bool     `tfsdk:"portal_customized"`
	Auth                               types.String   `tfsdk:"auth"`
	Expire                             types.Int64    `tfsdk:"expire"`
	ExpireNumber                       types.Int64    `tfsdk:"expire_number"`
	ExpireUnit                         types.Int64    `tfsdk:"expire_unit"`
	RedirectEnabled                    types.Bool     `tfsdk:"redirect_enabled"`
	RedirectHTTPS                      types.Bool     `tfsdk:"redirect_https"`
	RedirectToHTTPS                    types.Bool     `tfsdk:"redirect_to_https"`
	RestrictedSubnet1                  types.String   `tfsdk:"restricted_subnet_1"`
	RestrictedSubnet2                  types.String   `tfsdk:"restricted_subnet_2"`
	RestrictedSubnet3                  types.String   `tfsdk:"restricted_subnet_3"`
	PortalCustomizedTitle              types.String   `tfsdk:"portal_customized_title"`
	PortalCustomizedWelcomeText        types.String   `tfsdk:"portal_customized_welcome_text"`
	PortalCustomizedSuccessText        types.String   `tfsdk:"portal_customized_success_text"`
	PortalCustomizedAuthenticationText types.String   `tfsdk:"portal_customized_authentication_text"`
	PortalCustomizedButtonText         types.String   `tfsdk:"portal_customized_button_text"`
	PortalCustomizedButtonColor        types.String   `tfsdk:"portal_customized_button_color"`
	PortalCustomizedButtonTextColor    types.String   `tfsdk:"portal_customized_button_text_color"`
	PortalCustomizedLinkColor          types.String   `tfsdk:"portal_customized_link_color"`
	PortalCustomizedTextColor          types.String   `tfsdk:"portal_customized_text_color"`
	PortalCustomizedBgColor            types.String   `tfsdk:"portal_customized_bg_color"`
	PortalCustomizedBgType             types.String   `tfsdk:"portal_customized_bg_type"`
	PortalCustomizedBgImageEnabled     types.Bool     `tfsdk:"portal_customized_bg_image_enabled"`
	PortalCustomizedBoxColor           types.String   `tfsdk:"portal_customized_box_color"`
	PortalCustomizedBoxTextColor       types.String   `tfsdk:"portal_customized_box_text_color"`
	PortalCustomizedBoxOpacity         types.Int64    `tfsdk:"portal_customized_box_opacity"`
	PortalCustomizedBoxRadius          types.Int64    `tfsdk:"portal_customized_box_radius"`
	PortalCustomizedLogoPosition       types.String   `tfsdk:"portal_customized_logo_position"`
	PortalCustomizedLogoSize           types.Int64    `tfsdk:"portal_customized_logo_size"`
	PortalUseHostname                  types.Bool     `tfsdk:"portal_use_hostname"`
	ECEnabled                          types.Bool     `tfsdk:"ec_enabled"`
	TemplateEngine                     types.String   `tfsdk:"template_engine"`
	Timeouts                           timeouts.Value `tfsdk:"timeouts"`
}

func NewSettingGuestAccessResource() resource.Resource {
	return &SettingGuestAccessResource{}
}

func (r *SettingGuestAccessResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_guest_access"
}

func (r *SettingGuestAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi guest access/captive portal settings. " +
			"This is a singleton resource — one per site. Delete resets to defaults.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_enabled": schema.BoolAttribute{
				Description: "Enable the guest portal.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"portal_customized": schema.BoolAttribute{
				Description: "Enable portal customization.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"auth": schema.StringAttribute{
				Description: "Authentication type. Valid values: 'none', 'hotspot', 'facebook_wifi', 'custom'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "hotspot", "facebook_wifi", "custom"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expire": schema.Int64Attribute{
				Description: "Guest expiration in minutes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expire_number": schema.Int64Attribute{
				Description: "Number of expire units.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expire_unit": schema.Int64Attribute{
				Description: "Expire unit (1=hours, 60=minutes, 1440=days).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"redirect_enabled": schema.BoolAttribute{
				Description: "Enable redirect after authentication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"redirect_https": schema.BoolAttribute{
				Description: "Enable HTTPS redirect.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"redirect_to_https": schema.BoolAttribute{
				Description: "Redirect to HTTPS.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"restricted_subnet_1": schema.StringAttribute{
				Description: "First restricted subnet (CIDR).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"restricted_subnet_2": schema.StringAttribute{
				Description: "Second restricted subnet (CIDR).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"restricted_subnet_3": schema.StringAttribute{
				Description: "Third restricted subnet (CIDR).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_title": schema.StringAttribute{
				Description: "Portal page title.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_welcome_text": schema.StringAttribute{
				Description: "Portal welcome text.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_success_text": schema.StringAttribute{
				Description: "Portal success text.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_authentication_text": schema.StringAttribute{
				Description: "Portal authentication prompt text.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_button_text": schema.StringAttribute{
				Description: "Portal button text.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_button_color": schema.StringAttribute{
				Description: "Portal button color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_button_text_color": schema.StringAttribute{
				Description: "Portal button text color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_link_color": schema.StringAttribute{
				Description: "Portal link color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_text_color": schema.StringAttribute{
				Description: "Portal text color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_bg_color": schema.StringAttribute{
				Description: "Portal background color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_bg_type": schema.StringAttribute{
				Description: "Portal background type.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_bg_image_enabled": schema.BoolAttribute{
				Description: "Enable portal background image.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"portal_customized_box_color": schema.StringAttribute{
				Description: "Portal box color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_box_text_color": schema.StringAttribute{
				Description: "Portal box text color (hex).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_box_opacity": schema.Int64Attribute{
				Description: "Portal box opacity (0-100).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_box_radius": schema.Int64Attribute{
				Description: "Portal box border radius.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_logo_position": schema.StringAttribute{
				Description: "Portal logo position.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_customized_logo_size": schema.Int64Attribute{
				Description: "Portal logo size.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"portal_use_hostname": schema.BoolAttribute{
				Description: "Use hostname for portal.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ec_enabled": schema.BoolAttribute{
				Description: "Enable external captive portal.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"template_engine": schema.StringAttribute{
				Description: "Template engine. Valid values: 'angular', 'jsp'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("angular", "jsp"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *SettingGuestAccessResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingGuestAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingGuestAccessResourceModel
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

	setting := r.planToSDK(&plan)

	updated, err := r.client.UpdateSettingGuestAccess(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "guest access setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingGuestAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingGuestAccessResourceModel
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

	setting, err := r.client.GetSettingGuestAccess(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "guest access setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(setting, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingGuestAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingGuestAccessResourceModel
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

	setting := r.planToSDK(&plan)
	if !plan.ID.IsNull() {
		setting.ID = plan.ID.ValueString()
	}

	updated, err := r.client.UpdateSettingGuestAccess(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "guest access setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingGuestAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingGuestAccessResourceModel
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

	defaults := &unifi.SettingGuestAccess{
		Key:           "guest_access",
		PortalEnabled: boolPtr(false),
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingGuestAccess(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "guest access setting")
		return
	}
}

func (r *SettingGuestAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingGuestAccessResource) planToSDK(plan *SettingGuestAccessResourceModel) *unifi.SettingGuestAccess {
	s := &unifi.SettingGuestAccess{
		Key:                            "guest_access",
		PortalEnabled:                  boolPtr(plan.PortalEnabled.ValueBool()),
		PortalCustomized:               boolPtr(plan.PortalCustomized.ValueBool()),
		RedirectEnabled:                boolPtr(plan.RedirectEnabled.ValueBool()),
		RedirectHTTPS:                  boolPtr(plan.RedirectHTTPS.ValueBool()),
		RedirectToHTTPS:                boolPtr(plan.RedirectToHTTPS.ValueBool()),
		PortalCustomizedBgImageEnabled: boolPtr(plan.PortalCustomizedBgImageEnabled.ValueBool()),
		PortalUseHostname:              boolPtr(plan.PortalUseHostname.ValueBool()),
		ECEnabled:                      boolPtr(plan.ECEnabled.ValueBool()),
	}

	if !plan.Auth.IsNull() && !plan.Auth.IsUnknown() {
		s.Auth = plan.Auth.ValueString()
	}
	if !plan.Expire.IsNull() && !plan.Expire.IsUnknown() {
		v := int(plan.Expire.ValueInt64())
		s.Expire = &v
	}
	if !plan.ExpireNumber.IsNull() && !plan.ExpireNumber.IsUnknown() {
		v := int(plan.ExpireNumber.ValueInt64())
		s.ExpireNumber = &v
	}
	if !plan.ExpireUnit.IsNull() && !plan.ExpireUnit.IsUnknown() {
		v := int(plan.ExpireUnit.ValueInt64())
		s.ExpireUnit = &v
	}
	if !plan.RestrictedSubnet1.IsNull() && !plan.RestrictedSubnet1.IsUnknown() {
		s.RestrictedSubnet1 = plan.RestrictedSubnet1.ValueString()
	}
	if !plan.RestrictedSubnet2.IsNull() && !plan.RestrictedSubnet2.IsUnknown() {
		s.RestrictedSubnet2 = plan.RestrictedSubnet2.ValueString()
	}
	if !plan.RestrictedSubnet3.IsNull() && !plan.RestrictedSubnet3.IsUnknown() {
		s.RestrictedSubnet3 = plan.RestrictedSubnet3.ValueString()
	}
	if !plan.PortalCustomizedTitle.IsNull() && !plan.PortalCustomizedTitle.IsUnknown() {
		s.PortalCustomizedTitle = plan.PortalCustomizedTitle.ValueString()
	}
	if !plan.PortalCustomizedWelcomeText.IsNull() && !plan.PortalCustomizedWelcomeText.IsUnknown() {
		s.PortalCustomizedWelcomeText = plan.PortalCustomizedWelcomeText.ValueString()
	}
	if !plan.PortalCustomizedSuccessText.IsNull() && !plan.PortalCustomizedSuccessText.IsUnknown() {
		s.PortalCustomizedSuccessText = plan.PortalCustomizedSuccessText.ValueString()
	}
	if !plan.PortalCustomizedAuthenticationText.IsNull() && !plan.PortalCustomizedAuthenticationText.IsUnknown() {
		s.PortalCustomizedAuthenticationText = plan.PortalCustomizedAuthenticationText.ValueString()
	}
	if !plan.PortalCustomizedButtonText.IsNull() && !plan.PortalCustomizedButtonText.IsUnknown() {
		s.PortalCustomizedButtonText = plan.PortalCustomizedButtonText.ValueString()
	}
	if !plan.PortalCustomizedButtonColor.IsNull() && !plan.PortalCustomizedButtonColor.IsUnknown() {
		s.PortalCustomizedButtonColor = plan.PortalCustomizedButtonColor.ValueString()
	}
	if !plan.PortalCustomizedButtonTextColor.IsNull() && !plan.PortalCustomizedButtonTextColor.IsUnknown() {
		s.PortalCustomizedButtonTextColor = plan.PortalCustomizedButtonTextColor.ValueString()
	}
	if !plan.PortalCustomizedLinkColor.IsNull() && !plan.PortalCustomizedLinkColor.IsUnknown() {
		s.PortalCustomizedLinkColor = plan.PortalCustomizedLinkColor.ValueString()
	}
	if !plan.PortalCustomizedTextColor.IsNull() && !plan.PortalCustomizedTextColor.IsUnknown() {
		s.PortalCustomizedTextColor = plan.PortalCustomizedTextColor.ValueString()
	}
	if !plan.PortalCustomizedBgColor.IsNull() && !plan.PortalCustomizedBgColor.IsUnknown() {
		s.PortalCustomizedBgColor = plan.PortalCustomizedBgColor.ValueString()
	}
	if !plan.PortalCustomizedBgType.IsNull() && !plan.PortalCustomizedBgType.IsUnknown() {
		s.PortalCustomizedBgType = plan.PortalCustomizedBgType.ValueString()
	}
	if !plan.PortalCustomizedBoxColor.IsNull() && !plan.PortalCustomizedBoxColor.IsUnknown() {
		s.PortalCustomizedBoxColor = plan.PortalCustomizedBoxColor.ValueString()
	}
	if !plan.PortalCustomizedBoxTextColor.IsNull() && !plan.PortalCustomizedBoxTextColor.IsUnknown() {
		s.PortalCustomizedBoxTextColor = plan.PortalCustomizedBoxTextColor.ValueString()
	}
	if !plan.PortalCustomizedBoxOpacity.IsNull() && !plan.PortalCustomizedBoxOpacity.IsUnknown() {
		v := int(plan.PortalCustomizedBoxOpacity.ValueInt64())
		s.PortalCustomizedBoxOpacity = &v
	}
	if !plan.PortalCustomizedBoxRadius.IsNull() && !plan.PortalCustomizedBoxRadius.IsUnknown() {
		v := int(plan.PortalCustomizedBoxRadius.ValueInt64())
		s.PortalCustomizedBoxRadius = &v
	}
	if !plan.PortalCustomizedLogoPosition.IsNull() && !plan.PortalCustomizedLogoPosition.IsUnknown() {
		s.PortalCustomizedLogoPosition = plan.PortalCustomizedLogoPosition.ValueString()
	}
	if !plan.PortalCustomizedLogoSize.IsNull() && !plan.PortalCustomizedLogoSize.IsUnknown() {
		v := int(plan.PortalCustomizedLogoSize.ValueInt64())
		s.PortalCustomizedLogoSize = &v
	}
	if !plan.TemplateEngine.IsNull() && !plan.TemplateEngine.IsUnknown() {
		s.TemplateEngine = plan.TemplateEngine.ValueString()
	}

	return s
}

func (r *SettingGuestAccessResource) sdkToState(setting *unifi.SettingGuestAccess, state *SettingGuestAccessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.PortalEnabled = types.BoolValue(derefBool(setting.PortalEnabled))
	state.PortalCustomized = types.BoolValue(derefBool(setting.PortalCustomized))
	state.Auth = stringValueOrNull(setting.Auth)
	state.RedirectEnabled = types.BoolValue(derefBool(setting.RedirectEnabled))
	state.RedirectHTTPS = types.BoolValue(derefBool(setting.RedirectHTTPS))
	state.RedirectToHTTPS = types.BoolValue(derefBool(setting.RedirectToHTTPS))
	state.RestrictedSubnet1 = stringValueOrNull(setting.RestrictedSubnet1)
	state.RestrictedSubnet2 = stringValueOrNull(setting.RestrictedSubnet2)
	state.RestrictedSubnet3 = stringValueOrNull(setting.RestrictedSubnet3)
	state.PortalCustomizedTitle = stringValueOrNull(setting.PortalCustomizedTitle)
	state.PortalCustomizedWelcomeText = stringValueOrNull(setting.PortalCustomizedWelcomeText)
	state.PortalCustomizedSuccessText = stringValueOrNull(setting.PortalCustomizedSuccessText)
	state.PortalCustomizedAuthenticationText = stringValueOrNull(setting.PortalCustomizedAuthenticationText)
	state.PortalCustomizedButtonText = stringValueOrNull(setting.PortalCustomizedButtonText)
	state.PortalCustomizedButtonColor = stringValueOrNull(setting.PortalCustomizedButtonColor)
	state.PortalCustomizedButtonTextColor = stringValueOrNull(setting.PortalCustomizedButtonTextColor)
	state.PortalCustomizedLinkColor = stringValueOrNull(setting.PortalCustomizedLinkColor)
	state.PortalCustomizedTextColor = stringValueOrNull(setting.PortalCustomizedTextColor)
	state.PortalCustomizedBgColor = stringValueOrNull(setting.PortalCustomizedBgColor)
	state.PortalCustomizedBgType = stringValueOrNull(setting.PortalCustomizedBgType)
	state.PortalCustomizedBgImageEnabled = types.BoolValue(derefBool(setting.PortalCustomizedBgImageEnabled))
	state.PortalCustomizedBoxColor = stringValueOrNull(setting.PortalCustomizedBoxColor)
	state.PortalCustomizedBoxTextColor = stringValueOrNull(setting.PortalCustomizedBoxTextColor)
	state.PortalCustomizedLogoPosition = stringValueOrNull(setting.PortalCustomizedLogoPosition)
	state.PortalUseHostname = types.BoolValue(derefBool(setting.PortalUseHostname))
	state.ECEnabled = types.BoolValue(derefBool(setting.ECEnabled))
	state.TemplateEngine = stringValueOrNull(setting.TemplateEngine)

	if setting.Expire != nil {
		state.Expire = types.Int64Value(int64(*setting.Expire))
	} else {
		state.Expire = types.Int64Null()
	}
	if setting.ExpireNumber != nil {
		state.ExpireNumber = types.Int64Value(int64(*setting.ExpireNumber))
	} else {
		state.ExpireNumber = types.Int64Null()
	}
	if setting.ExpireUnit != nil {
		state.ExpireUnit = types.Int64Value(int64(*setting.ExpireUnit))
	} else {
		state.ExpireUnit = types.Int64Null()
	}
	if setting.PortalCustomizedBoxOpacity != nil {
		state.PortalCustomizedBoxOpacity = types.Int64Value(int64(*setting.PortalCustomizedBoxOpacity))
	} else {
		state.PortalCustomizedBoxOpacity = types.Int64Null()
	}
	if setting.PortalCustomizedBoxRadius != nil {
		state.PortalCustomizedBoxRadius = types.Int64Value(int64(*setting.PortalCustomizedBoxRadius))
	} else {
		state.PortalCustomizedBoxRadius = types.Int64Null()
	}
	if setting.PortalCustomizedLogoSize != nil {
		state.PortalCustomizedLogoSize = types.Int64Value(int64(*setting.PortalCustomizedLogoSize))
	} else {
		state.PortalCustomizedLogoSize = types.Int64Null()
	}

	return diags
}
