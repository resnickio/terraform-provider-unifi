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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &WLANResource{}
	_ resource.ResourceWithImportState = &WLANResource{}
)

type WLANResource struct {
	client *AutoLoginClient
}

type WLANResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	SiteID           types.String   `tfsdk:"site_id"`
	Name             types.String   `tfsdk:"name"`
	Enabled          types.Bool     `tfsdk:"enabled"`
	Security         types.String   `tfsdk:"security"`
	WPAMode          types.String   `tfsdk:"wpa_mode"`
	WPAEnc           types.String   `tfsdk:"wpa_enc"`
	Passphrase       types.String   `tfsdk:"passphrase"`
	NetworkID        types.String   `tfsdk:"network_id"`
	UserGroupID      types.String   `tfsdk:"user_group_id"`
	APGroupIDs       types.Set      `tfsdk:"ap_group_ids"`
	IsGuest          types.Bool     `tfsdk:"is_guest"`
	HideSsid         types.Bool     `tfsdk:"hide_ssid"`
	WLANBand         types.String   `tfsdk:"wlan_band"`
	WLANBands        types.Set      `tfsdk:"wlan_bands"`
	Vlan             types.Int64    `tfsdk:"vlan"`
	VlanEnabled      types.Bool     `tfsdk:"vlan_enabled"`
	MacFilterEnabled types.Bool     `tfsdk:"mac_filter_enabled"`
	MacFilterList    types.Set      `tfsdk:"mac_filter_list"`
	MacFilterPolicy  types.String   `tfsdk:"mac_filter_policy"`
	ScheduleEnabled  types.Bool     `tfsdk:"schedule_enabled"`
	Schedule         types.Set      `tfsdk:"schedule"`
	L2Isolation      types.Bool     `tfsdk:"l2_isolation"`
	FastRoaming      types.Bool     `tfsdk:"fast_roaming_enabled"`
	ProxyArp         types.Bool     `tfsdk:"proxy_arp"`
	BssTransition    types.Bool     `tfsdk:"bss_transition"`
	Uapsd            types.Bool     `tfsdk:"uapsd_enabled"`
	PmfMode          types.String   `tfsdk:"pmf_mode"`
	WPA3Support      types.Bool     `tfsdk:"wpa3_support"`
	WPA3Transition   types.Bool     `tfsdk:"wpa3_transition"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func NewWLANResource() resource.Resource {
	return &WLANResource{}
}

func (r *WLANResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wlan"
}

func (r *WLANResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi wireless network (SSID) configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the WLAN.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the WLAN is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The SSID name of the wireless network.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the WLAN is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"security": schema.StringAttribute{
				Description: "The security mode. Valid values: 'open', 'wep', 'wpapsk', 'wpaeap'. Defaults to 'wpapsk'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("wpapsk"),
				Validators: []validator.String{
					stringvalidator.OneOf("open", "wep", "wpapsk", "wpaeap"),
				},
			},
			"wpa_mode": schema.StringAttribute{
				Description: "The WPA mode. Valid values: 'auto', 'wpa1', 'wpa2'. Defaults to 'wpa2'. Note: WPA3 is controlled via wpa3_support attribute.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("wpa2"),
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "wpa1", "wpa2"),
				},
			},
			"wpa_enc": schema.StringAttribute{
				Description: "The WPA encryption type. Valid values: 'ccmp', 'gcmp', 'auto'. Defaults to 'ccmp'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ccmp"),
			},
			"passphrase": schema.StringAttribute{
				Description: "The wireless passphrase (required for wpapsk security).",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID (VLAN) to assign to this WLAN.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_group_id": schema.StringAttribute{
				Description: "The user group ID for bandwidth limiting.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ap_group_ids": schema.SetAttribute{
				Description: "Set of AP group IDs this WLAN should be broadcast on. Required.",
				Required:    true,
				ElementType: types.StringType,
			},
			"is_guest": schema.BoolAttribute{
				Description: "Whether this is a guest network. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"hide_ssid": schema.BoolAttribute{
				Description: "Whether to hide the SSID. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"wlan_band": schema.StringAttribute{
				Description: "The wireless band. Valid values: '2g', '5g', 'both'. Defaults to 'both'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("both"),
			},
			"wlan_bands": schema.SetAttribute{
				Description: "Set of wireless bands to enable (e.g., '2g', '5g', '6g').",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"vlan": schema.Int64Attribute{
				Description: "The VLAN ID for this WLAN. Note: On modern UniFi controllers (v8+), VLAN tagging is done by associating the WLAN with a network that has the desired VLAN ID.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"vlan_enabled": schema.BoolAttribute{
				Description: "Whether VLAN tagging is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"mac_filter_enabled": schema.BoolAttribute{
				Description: "Whether MAC filtering is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"mac_filter_list": schema.SetAttribute{
				Description: "Set of MAC addresses for filtering.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mac_filter_policy": schema.StringAttribute{
				Description: "MAC filter policy. Valid values: 'allow', 'deny'. Defaults to 'deny'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("deny"),
				Validators: []validator.String{
					stringvalidator.OneOf("allow", "deny"),
				},
			},
			"schedule_enabled": schema.BoolAttribute{
				Description: "Whether scheduling is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"schedule": schema.SetAttribute{
				Description: "Schedule configuration.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"l2_isolation": schema.BoolAttribute{
				Description: "Whether L2 isolation is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"fast_roaming_enabled": schema.BoolAttribute{
				Description: "Whether fast roaming (802.11r) is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"proxy_arp": schema.BoolAttribute{
				Description: "Whether proxy ARP is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"bss_transition": schema.BoolAttribute{
				Description: "Whether BSS transition is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"uapsd_enabled": schema.BoolAttribute{
				Description: "Whether U-APSD (WMM Power Save) is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"pmf_mode": schema.StringAttribute{
				Description: "Protected Management Frames mode. Valid values: 'disabled', 'optional', 'required'. Defaults to 'optional'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("optional"),
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "optional", "required"),
				},
			},
			"wpa3_support": schema.BoolAttribute{
				Description: "Whether WPA3 support is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"wpa3_transition": schema.BoolAttribute{
				Description: "Whether WPA3 transition mode is enabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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

func (r *WLANResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WLANResourceModel

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

	wlan := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateWLAN(ctx, wlan)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "WLAN")
		return
	}

	originalPassphrase := plan.Passphrase

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !originalPassphrase.IsNull() && plan.Passphrase.IsNull() {
		plan.Passphrase = originalPassphrase
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WLANResourceModel

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

	priorState := state

	wlan, err := r.client.GetWLAN(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "WLAN")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, wlan, &state, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WLANResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WLANResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WLANResourceModel
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

	wlan := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	wlan.ID = state.ID.ValueString()
	wlan.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateWLAN(ctx, state.ID.ValueString(), wlan)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "WLAN")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WLANResourceModel

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

	err := r.client.DeleteWLAN(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "WLAN")
		return
	}
}

func (r *WLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *WLANResource) planToSDK(ctx context.Context, plan *WLANResourceModel, diags *diag.Diagnostics) *unifi.WLANConf {
	wlan := &unifi.WLANConf{
		Name:               plan.Name.ValueString(),
		Enabled:            boolPtr(plan.Enabled.ValueBool()),
		Security:           plan.Security.ValueString(),
		WPAMode:            plan.WPAMode.ValueString(),
		WPAEnc:             plan.WPAEnc.ValueString(),
		IsGuest:            boolPtr(plan.IsGuest.ValueBool()),
		HideSsid:           boolPtr(plan.HideSsid.ValueBool()),
		WLANBand:           plan.WLANBand.ValueString(),
		VlanEnabled:        boolPtr(plan.VlanEnabled.ValueBool()),
		MacFilterEnabled:   boolPtr(plan.MacFilterEnabled.ValueBool()),
		MacFilterPolicy:    plan.MacFilterPolicy.ValueString(),
		ScheduleEnabled:    boolPtr(plan.ScheduleEnabled.ValueBool()),
		L2Isolation:        boolPtr(plan.L2Isolation.ValueBool()),
		FastRoamingEnabled: boolPtr(plan.FastRoaming.ValueBool()),
		ProxyArp:           boolPtr(plan.ProxyArp.ValueBool()),
		BssTransition:      boolPtr(plan.BssTransition.ValueBool()),
		Uapsd:              boolPtr(plan.Uapsd.ValueBool()),
		PmfMode:            plan.PmfMode.ValueString(),
		WPA3Support:        boolPtr(plan.WPA3Support.ValueBool()),
		WPA3Transition:     boolPtr(plan.WPA3Transition.ValueBool()),
	}

	if !plan.Passphrase.IsNull() && !plan.Passphrase.IsUnknown() {
		wlan.XPassphrase = plan.Passphrase.ValueString()
	}

	if !plan.NetworkID.IsNull() && !plan.NetworkID.IsUnknown() {
		wlan.NetworkConfID = plan.NetworkID.ValueString()
	}

	if !plan.UserGroupID.IsNull() && !plan.UserGroupID.IsUnknown() {
		wlan.Usergroup = plan.UserGroupID.ValueString()
	}

	if !plan.APGroupIDs.IsNull() && !plan.APGroupIDs.IsUnknown() {
		var apGroupIDs []string
		diags.Append(plan.APGroupIDs.ElementsAs(ctx, &apGroupIDs, false)...)
		if diags.HasError() {
			return nil
		}
		wlan.APGroupIDs = apGroupIDs
	}

	if !plan.Vlan.IsNull() && !plan.Vlan.IsUnknown() {
		wlan.Vlan = intPtr(plan.Vlan.ValueInt64())
	}

	if !plan.WLANBands.IsNull() && !plan.WLANBands.IsUnknown() {
		var bands []string
		diags.Append(plan.WLANBands.ElementsAs(ctx, &bands, false)...)
		if diags.HasError() {
			return nil
		}
		wlan.WLANBands = bands
	}

	if !plan.MacFilterList.IsNull() && !plan.MacFilterList.IsUnknown() {
		var macList []string
		diags.Append(plan.MacFilterList.ElementsAs(ctx, &macList, false)...)
		if diags.HasError() {
			return nil
		}
		wlan.MacFilterList = macList
	}

	if !plan.Schedule.IsNull() && !plan.Schedule.IsUnknown() {
		var schedule []string
		diags.Append(plan.Schedule.ElementsAs(ctx, &schedule, false)...)
		if diags.HasError() {
			return nil
		}
		wlan.Schedule = schedule
	}

	return wlan
}

func (r *WLANResource) sdkToState(ctx context.Context, wlan *unifi.WLANConf, state *WLANResourceModel, priorState *WLANResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(wlan.ID)
	state.SiteID = types.StringValue(wlan.SiteID)
	state.Name = types.StringValue(wlan.Name)
	state.Enabled = types.BoolValue(derefBool(wlan.Enabled))
	state.Security = types.StringValue(wlan.Security)
	state.WPAMode = types.StringValue(wlan.WPAMode)
	state.WPAEnc = types.StringValue(wlan.WPAEnc)

	if wlan.XPassphrase != "" {
		state.Passphrase = types.StringValue(wlan.XPassphrase)
	} else if priorState != nil && !priorState.Passphrase.IsNull() {
		state.Passphrase = priorState.Passphrase
	} else {
		state.Passphrase = types.StringNull()
	}

	if wlan.NetworkConfID != "" {
		state.NetworkID = types.StringValue(wlan.NetworkConfID)
	} else {
		state.NetworkID = types.StringNull()
	}

	if wlan.Usergroup != "" {
		state.UserGroupID = types.StringValue(wlan.Usergroup)
	} else {
		state.UserGroupID = types.StringNull()
	}

	if len(wlan.APGroupIDs) > 0 {
		apGroupSet, d := types.SetValueFrom(ctx, types.StringType, wlan.APGroupIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.APGroupIDs = apGroupSet
	} else {
		state.APGroupIDs = types.SetNull(types.StringType)
	}

	state.IsGuest = types.BoolValue(derefBool(wlan.IsGuest))
	state.HideSsid = types.BoolValue(derefBool(wlan.HideSsid))
	state.WLANBand = types.StringValue(wlan.WLANBand)

	if len(wlan.WLANBands) > 0 {
		bandsSet, d := types.SetValueFrom(ctx, types.StringType, wlan.WLANBands)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.WLANBands = bandsSet
	} else {
		state.WLANBands = types.SetNull(types.StringType)
	}

	if wlan.Vlan != nil {
		state.Vlan = types.Int64Value(int64(*wlan.Vlan))
	} else {
		state.Vlan = types.Int64Null()
	}

	state.VlanEnabled = types.BoolValue(derefBool(wlan.VlanEnabled))
	state.MacFilterEnabled = types.BoolValue(derefBool(wlan.MacFilterEnabled))

	if len(wlan.MacFilterList) > 0 {
		macSet, d := types.SetValueFrom(ctx, types.StringType, wlan.MacFilterList)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.MacFilterList = macSet
	} else {
		state.MacFilterList = types.SetNull(types.StringType)
	}

	state.MacFilterPolicy = types.StringValue(wlan.MacFilterPolicy)
	state.ScheduleEnabled = types.BoolValue(derefBool(wlan.ScheduleEnabled))

	if len(wlan.Schedule) > 0 {
		scheduleSet, d := types.SetValueFrom(ctx, types.StringType, wlan.Schedule)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Schedule = scheduleSet
	} else {
		state.Schedule = types.SetNull(types.StringType)
	}

	state.L2Isolation = types.BoolValue(derefBool(wlan.L2Isolation))
	state.FastRoaming = types.BoolValue(derefBool(wlan.FastRoamingEnabled))
	state.ProxyArp = types.BoolValue(derefBool(wlan.ProxyArp))
	state.BssTransition = types.BoolValue(derefBool(wlan.BssTransition))
	state.Uapsd = types.BoolValue(derefBool(wlan.Uapsd))
	state.PmfMode = types.StringValue(wlan.PmfMode)
	state.WPA3Support = types.BoolValue(derefBool(wlan.WPA3Support))
	state.WPA3Transition = types.BoolValue(derefBool(wlan.WPA3Transition))

	return diags
}
