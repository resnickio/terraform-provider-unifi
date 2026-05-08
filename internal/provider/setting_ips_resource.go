package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &SettingIPSResource{}
	_ resource.ResourceWithImportState = &SettingIPSResource{}
)

type SettingIPSResource struct {
	client *AutoLoginClient
}

type SettingIPSResourceModel struct {
	ID                                  types.String   `tfsdk:"id"`
	SiteID                              types.String   `tfsdk:"site_id"`
	IPSMode                             types.String   `tfsdk:"ips_mode"`
	DNSFiltering                        types.Bool     `tfsdk:"dns_filtering"`
	DNSFilters                          types.List     `tfsdk:"dns_filters"`
	EnabledCategories                   types.Set      `tfsdk:"enabled_categories"`
	HoneypotEnabled                     types.Bool     `tfsdk:"honeypot_enabled"`
	EndpointScanning                    types.Bool     `tfsdk:"endpoint_scanning"`
	AdBlockingEnabled                   types.Bool     `tfsdk:"ad_blocking_enabled"`
	AdvancedFilteringPreference         types.String   `tfsdk:"advanced_filtering_preference"`
	ContentFilteringBlockingPageEnabled types.Bool     `tfsdk:"content_filtering_blocking_page_enabled"`
	MemoryOptimized                     types.Bool     `tfsdk:"memory_optimized"`
	SuppressionAlerts                   types.String   `tfsdk:"suppression_alerts"`
	SuppressionWhitelist                types.String   `tfsdk:"suppression_whitelist"`
	Timeouts                            timeouts.Value `tfsdk:"timeouts"`
}

type DNSFilterModel struct {
	Filter       types.String `tfsdk:"filter"`
	NetworkID    types.String `tfsdk:"network_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Version      types.String `tfsdk:"version"`
	BlockedTLD   types.Set    `tfsdk:"blocked_tld"`
	BlockedSites types.Set    `tfsdk:"blocked_sites"`
	AllowedSites types.Set    `tfsdk:"allowed_sites"`
}

var dnsFilterAttrTypes = map[string]attr.Type{
	"filter":        types.StringType,
	"network_id":    types.StringType,
	"name":          types.StringType,
	"description":   types.StringType,
	"version":       types.StringType,
	"blocked_tld":   types.SetType{ElemType: types.StringType},
	"blocked_sites": types.SetType{ElemType: types.StringType},
	"allowed_sites": types.SetType{ElemType: types.StringType},
}

func NewSettingIPSResource() resource.Resource {
	return &SettingIPSResource{}
}

func (r *SettingIPSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_ips"
}

func (r *SettingIPSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi IPS/IDS and threat management settings. " +
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
			"ips_mode": schema.StringAttribute{
				Description: "IPS mode. Valid values: 'disabled', 'ids', 'ips'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "ids", "ips"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns_filtering": schema.BoolAttribute{
				Description: "Enable DNS filtering.",
				Optional:    true,
				Computed:    true,
			},
			"dns_filters": schema.ListNestedAttribute{
				Description: "DNS filter configurations.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"filter": schema.StringAttribute{
							Description: "Filter type/identifier.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"network_id": schema.StringAttribute{
							Description: "Network ID this filter applies to.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Description: "Filter name.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"description": schema.StringAttribute{
							Description: "Filter description.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"version": schema.StringAttribute{
							Description: "Filter version.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"blocked_tld": schema.SetAttribute{
							Description: "Blocked top-level domains.",
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
						},
						"blocked_sites": schema.SetAttribute{
							Description: "Blocked sites.",
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
						},
						"allowed_sites": schema.SetAttribute{
							Description: "Allowed sites (overrides).",
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"enabled_categories": schema.SetAttribute{
				Description: "Set of enabled IPS categories.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"honeypot_enabled": schema.BoolAttribute{
				Description: "Enable honeypot.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"endpoint_scanning": schema.BoolAttribute{
				Description: "Enable endpoint scanning.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ad_blocking_enabled": schema.BoolAttribute{
				Description: "Enable ad blocking.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"advanced_filtering_preference": schema.StringAttribute{
				Description: "Advanced filtering preference. Valid values: 'disabled', 'manual'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "manual"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content_filtering_blocking_page_enabled": schema.BoolAttribute{
				Description: "Enable content filtering blocking page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"memory_optimized": schema.BoolAttribute{
				Description: "Enable memory-optimized mode.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"suppression_alerts": schema.StringAttribute{
				Description: "IPS suppression alerts as JSON string.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"suppression_whitelist": schema.StringAttribute{
				Description: "IPS suppression whitelist as JSON string.",
				Optional:    true,
				Computed:    true,
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

func (r *SettingIPSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingIPSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingIPSResourceModel
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

	setting := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateSettingIPS(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "IPS setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingIPSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingIPSResourceModel
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

	setting, err := r.client.GetSettingIPS(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "IPS setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, setting, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingIPSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingIPSResourceModel
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

	setting := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.ID.IsNull() {
		setting.ID = plan.ID.ValueString()
	}

	updated, err := r.client.UpdateSettingIPS(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "IPS setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingIPSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingIPSResourceModel
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

	defaults := &unifi.SettingIPS{
		Key:     "ips",
		IPSMode: "disabled",
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingIPS(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "IPS setting")
		return
	}
}

func (r *SettingIPSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingIPSResource) planToSDK(ctx context.Context, plan *SettingIPSResourceModel, diags *diag.Diagnostics) *unifi.SettingIPS {
	s := &unifi.SettingIPS{
		Key:                                "ips",
		DNSFiltering:                       boolPtr(plan.DNSFiltering.ValueBool()),
		HoneypotEnabled:                    boolPtr(plan.HoneypotEnabled.ValueBool()),
		EndpointScanning:                   boolPtr(plan.EndpointScanning.ValueBool()),
		AdBlockingEnabled:                  boolPtr(plan.AdBlockingEnabled.ValueBool()),
		ContentFilteringBlockingPageEnable: boolPtr(plan.ContentFilteringBlockingPageEnabled.ValueBool()),
		MemoryOptimized:                    boolPtr(plan.MemoryOptimized.ValueBool()),
	}

	if !plan.IPSMode.IsNull() && !plan.IPSMode.IsUnknown() {
		s.IPSMode = plan.IPSMode.ValueString()
	}
	if !plan.AdvancedFilteringPreference.IsNull() && !plan.AdvancedFilteringPreference.IsUnknown() {
		s.AdvancedFilteringPreference = plan.AdvancedFilteringPreference.ValueString()
	}

	if !plan.EnabledCategories.IsNull() && !plan.EnabledCategories.IsUnknown() {
		var categories []string
		diags.Append(plan.EnabledCategories.ElementsAs(ctx, &categories, false)...)
		s.EnabledCategories = categories
	}

	if !plan.DNSFilters.IsNull() && !plan.DNSFilters.IsUnknown() {
		var filters []DNSFilterModel
		diags.Append(plan.DNSFilters.ElementsAs(ctx, &filters, false)...)
		for _, f := range filters {
			sdkFilter := unifi.DNSFilter{
				Filter:      f.Filter.ValueString(),
				NetworkID:   f.NetworkID.ValueString(),
				Name:        f.Name.ValueString(),
				Description: f.Description.ValueString(),
				Version:     f.Version.ValueString(),
			}
			if !f.BlockedTLD.IsNull() && !f.BlockedTLD.IsUnknown() {
				var tlds []string
				diags.Append(f.BlockedTLD.ElementsAs(ctx, &tlds, false)...)
				sdkFilter.BlockedTLD = tlds
			}
			if !f.BlockedSites.IsNull() && !f.BlockedSites.IsUnknown() {
				var sites []string
				diags.Append(f.BlockedSites.ElementsAs(ctx, &sites, false)...)
				sdkFilter.BlockedSites = sites
			}
			if !f.AllowedSites.IsNull() && !f.AllowedSites.IsUnknown() {
				var sites []string
				diags.Append(f.AllowedSites.ElementsAs(ctx, &sites, false)...)
				sdkFilter.AllowedSites = sites
			}
			s.DNSFilters = append(s.DNSFilters, sdkFilter)
		}
	}

	if !plan.SuppressionAlerts.IsNull() && !plan.SuppressionAlerts.IsUnknown() {
		if s.Suppression == nil {
			s.Suppression = &unifi.IPSSuppression{}
		}
		var alerts []json.RawMessage
		if err := json.Unmarshal([]byte(plan.SuppressionAlerts.ValueString()), &alerts); err != nil {
			diags.AddError("Invalid suppression_alerts JSON", err.Error())
		} else {
			s.Suppression.Alerts = alerts
		}
	}
	if !plan.SuppressionWhitelist.IsNull() && !plan.SuppressionWhitelist.IsUnknown() {
		if s.Suppression == nil {
			s.Suppression = &unifi.IPSSuppression{}
		}
		var whitelist []json.RawMessage
		if err := json.Unmarshal([]byte(plan.SuppressionWhitelist.ValueString()), &whitelist); err != nil {
			diags.AddError("Invalid suppression_whitelist JSON", err.Error())
		} else {
			s.Suppression.Whitelist = whitelist
		}
	}

	return s
}

func (r *SettingIPSResource) sdkToState(ctx context.Context, setting *unifi.SettingIPS, state *SettingIPSResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.IPSMode = stringValueOrNull(setting.IPSMode)
	state.DNSFiltering = types.BoolValue(derefBool(setting.DNSFiltering))
	state.HoneypotEnabled = types.BoolValue(derefBool(setting.HoneypotEnabled))
	state.EndpointScanning = types.BoolValue(derefBool(setting.EndpointScanning))
	state.AdBlockingEnabled = types.BoolValue(derefBool(setting.AdBlockingEnabled))
	state.AdvancedFilteringPreference = stringValueOrNull(setting.AdvancedFilteringPreference)
	state.ContentFilteringBlockingPageEnabled = types.BoolValue(derefBool(setting.ContentFilteringBlockingPageEnable))
	state.MemoryOptimized = types.BoolValue(derefBool(setting.MemoryOptimized))

	if len(setting.EnabledCategories) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, setting.EnabledCategories)
		diags.Append(d...)
		state.EnabledCategories = s
	} else {
		state.EnabledCategories = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if len(setting.DNSFilters) > 0 {
		filterValues := make([]attr.Value, len(setting.DNSFilters))
		for i, f := range setting.DNSFilters {
			blockedTLD := types.SetValueMust(types.StringType, []attr.Value{})
			if len(f.BlockedTLD) > 0 {
				s, d := types.SetValueFrom(ctx, types.StringType, f.BlockedTLD)
				diags.Append(d...)
				blockedTLD = s
			}
			blockedSites := types.SetValueMust(types.StringType, []attr.Value{})
			if len(f.BlockedSites) > 0 {
				s, d := types.SetValueFrom(ctx, types.StringType, f.BlockedSites)
				diags.Append(d...)
				blockedSites = s
			}
			allowedSites := types.SetValueMust(types.StringType, []attr.Value{})
			if len(f.AllowedSites) > 0 {
				s, d := types.SetValueFrom(ctx, types.StringType, f.AllowedSites)
				diags.Append(d...)
				allowedSites = s
			}

			obj, d := types.ObjectValue(dnsFilterAttrTypes, map[string]attr.Value{
				"filter":        stringValueOrNull(f.Filter),
				"network_id":    stringValueOrNull(f.NetworkID),
				"name":          stringValueOrNull(f.Name),
				"description":   stringValueOrNull(f.Description),
				"version":       stringValueOrNull(f.Version),
				"blocked_tld":   blockedTLD,
				"blocked_sites": blockedSites,
				"allowed_sites": allowedSites,
			})
			diags.Append(d...)
			filterValues[i] = obj
		}
		list, d := types.ListValue(types.ObjectType{AttrTypes: dnsFilterAttrTypes}, filterValues)
		diags.Append(d...)
		state.DNSFilters = list
	} else {
		state.DNSFilters = types.ListValueMust(types.ObjectType{AttrTypes: dnsFilterAttrTypes}, []attr.Value{})
	}

	if setting.Suppression != nil && len(setting.Suppression.Alerts) > 0 {
		data, err := json.Marshal(setting.Suppression.Alerts)
		if err != nil {
			diags.AddError("Failed to serialize suppression_alerts", err.Error())
		} else {
			state.SuppressionAlerts = types.StringValue(string(data))
		}
	} else {
		state.SuppressionAlerts = types.StringNull()
	}

	if setting.Suppression != nil && len(setting.Suppression.Whitelist) > 0 {
		data, err := json.Marshal(setting.Suppression.Whitelist)
		if err != nil {
			diags.AddError("Failed to serialize suppression_whitelist", err.Error())
		} else {
			state.SuppressionWhitelist = types.StringValue(string(data))
		}
	} else {
		state.SuppressionWhitelist = types.StringNull()
	}

	return diags
}
