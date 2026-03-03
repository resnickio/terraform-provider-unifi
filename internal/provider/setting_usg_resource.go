package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &SettingUSGResource{}
	_ resource.ResourceWithImportState = &SettingUSGResource{}
)

type SettingUSGResource struct {
	client *AutoLoginClient
}

type SettingUSGResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	SiteID            types.String   `tfsdk:"site_id"`
	BroadcastPing     types.Bool     `tfsdk:"broadcast_ping"`
	DHCPDUseDnsmasq   types.Bool     `tfsdk:"dhcpd_use_dnsmasq"`
	FTPModule         types.Bool     `tfsdk:"ftp_module"`
	GREModule         types.Bool     `tfsdk:"gre_module"`
	H323Module        types.Bool     `tfsdk:"h323_module"`
	LLDPEnableAll     types.Bool     `tfsdk:"lldp_enable_all"`
	MDNSEnabled       types.Bool     `tfsdk:"mdns_enabled"`
	MSSClamp          types.String   `tfsdk:"mss_clamp"`
	MSSClampMSS       types.Int64    `tfsdk:"mss_clamp_mss"`
	OffloadAccounting types.Bool     `tfsdk:"offload_accounting"`
	OffloadL2Blocking types.Bool     `tfsdk:"offload_l2_blocking"`
	OffloadSch        types.Bool     `tfsdk:"offload_sch"`
	PPTPModule        types.Bool     `tfsdk:"pptp_module"`
	ReceiveRedirects  types.Bool     `tfsdk:"receive_redirects"`
	SendRedirects     types.Bool     `tfsdk:"send_redirects"`
	SIPModule         types.Bool     `tfsdk:"sip_module"`
	SynCookies        types.Bool     `tfsdk:"syn_cookies"`
	TFTPModule        types.Bool     `tfsdk:"tftp_module"`
	UPnPEnabled       types.Bool     `tfsdk:"upnp_enabled"`
	UPnPNATPMPEnabled types.Bool     `tfsdk:"upnp_nat_pmp_enabled"`
	UPnPSecureMode    types.Bool     `tfsdk:"upnp_secure_mode"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func NewSettingUSGResource() resource.Resource {
	return &SettingUSGResource{}
}

func (r *SettingUSGResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting_usg"
}

func (r *SettingUSGResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages UniFi gateway/security settings. Singleton resource — one per site. Delete resets to defaults.",
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
			"broadcast_ping": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"dhcpd_use_dnsmasq": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"ftp_module": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"gre_module": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"h323_module": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"lldp_enable_all": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"mdns_enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"mss_clamp": schema.StringAttribute{
				Description: "MSS clamping mode: 'auto' or 'custom'.",
				Optional:    true,
				Computed:    true,
			},
			"mss_clamp_mss": schema.Int64Attribute{
				Description: "MSS clamp value (when mss_clamp is 'custom').",
				Optional:    true,
				Computed:    true,
			},
			"offload_accounting": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"offload_l2_blocking": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"offload_sch": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"pptp_module": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"receive_redirects": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"send_redirects": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"sip_module": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"syn_cookies": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"tftp_module": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"upnp_enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"upnp_nat_pmp_enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"upnp_secure_mode": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
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

func (r *SettingUSGResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SettingUSGResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SettingUSGResourceModel
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
	updated, err := r.client.UpdateSettingUSG(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "USG setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingUSGResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingUSGResourceModel
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

	setting, err := r.client.GetSettingUSG(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "read", "USG setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(setting, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SettingUSGResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingUSGResourceModel
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

	updated, err := r.client.UpdateSettingUSG(ctx, setting)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "USG setting")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SettingUSGResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SettingUSGResourceModel
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

	defaults := &unifi.SettingUSG{
		BroadcastPing:     boolPtr(false),
		DHCPDUseDnsmasq:   boolPtr(false),
		FTPModule:         boolPtr(false),
		GREModule:         boolPtr(false),
		H323Module:        boolPtr(false),
		LLDPEnableAll:     boolPtr(false),
		MDNSEnabled:       boolPtr(false),
		OffloadAccounting: boolPtr(false),
		OffloadL2Blocking: boolPtr(false),
		OffloadSch:        boolPtr(false),
		PPTPModule:        boolPtr(false),
		ReceiveRedirects:  boolPtr(false),
		SendRedirects:     boolPtr(false),
		SIPModule:         boolPtr(false),
		SynCookies:        boolPtr(false),
		TFTPModule:        boolPtr(false),
		UPnPEnabled:       boolPtr(false),
		UPnPNATPMPEnabled: boolPtr(false),
		UPnPSecureMode:    boolPtr(false),
	}
	if !state.ID.IsNull() {
		defaults.ID = state.ID.ValueString()
	}

	_, err := r.client.UpdateSettingUSG(ctx, defaults)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "reset", "USG setting")
	}
}

func (r *SettingUSGResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SettingUSGResource) planToSDK(plan *SettingUSGResourceModel) *unifi.SettingUSG {
	s := &unifi.SettingUSG{
		Key:               "usg",
		BroadcastPing:     boolPtr(plan.BroadcastPing.ValueBool()),
		DHCPDUseDnsmasq:   boolPtr(plan.DHCPDUseDnsmasq.ValueBool()),
		FTPModule:         boolPtr(plan.FTPModule.ValueBool()),
		GREModule:         boolPtr(plan.GREModule.ValueBool()),
		H323Module:        boolPtr(plan.H323Module.ValueBool()),
		LLDPEnableAll:     boolPtr(plan.LLDPEnableAll.ValueBool()),
		MDNSEnabled:       boolPtr(plan.MDNSEnabled.ValueBool()),
		OffloadAccounting: boolPtr(plan.OffloadAccounting.ValueBool()),
		OffloadL2Blocking: boolPtr(plan.OffloadL2Blocking.ValueBool()),
		OffloadSch:        boolPtr(plan.OffloadSch.ValueBool()),
		PPTPModule:        boolPtr(plan.PPTPModule.ValueBool()),
		ReceiveRedirects:  boolPtr(plan.ReceiveRedirects.ValueBool()),
		SendRedirects:     boolPtr(plan.SendRedirects.ValueBool()),
		SIPModule:         boolPtr(plan.SIPModule.ValueBool()),
		SynCookies:        boolPtr(plan.SynCookies.ValueBool()),
		TFTPModule:        boolPtr(plan.TFTPModule.ValueBool()),
		UPnPEnabled:       boolPtr(plan.UPnPEnabled.ValueBool()),
		UPnPNATPMPEnabled: boolPtr(plan.UPnPNATPMPEnabled.ValueBool()),
		UPnPSecureMode:    boolPtr(plan.UPnPSecureMode.ValueBool()),
	}

	if !plan.MSSClamp.IsNull() && !plan.MSSClamp.IsUnknown() {
		s.MSSClamp = plan.MSSClamp.ValueString()
	}
	if !plan.MSSClampMSS.IsNull() && !plan.MSSClampMSS.IsUnknown() {
		v := int(plan.MSSClampMSS.ValueInt64())
		s.MSSClampMSS = &v
	}

	return s
}

func (r *SettingUSGResource) sdkToState(setting *unifi.SettingUSG, state *SettingUSGResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(setting.ID)
	state.SiteID = types.StringValue(setting.SiteID)
	state.BroadcastPing = types.BoolValue(derefBool(setting.BroadcastPing))
	state.DHCPDUseDnsmasq = types.BoolValue(derefBool(setting.DHCPDUseDnsmasq))
	state.FTPModule = types.BoolValue(derefBool(setting.FTPModule))
	state.GREModule = types.BoolValue(derefBool(setting.GREModule))
	state.H323Module = types.BoolValue(derefBool(setting.H323Module))
	state.LLDPEnableAll = types.BoolValue(derefBool(setting.LLDPEnableAll))
	state.MDNSEnabled = types.BoolValue(derefBool(setting.MDNSEnabled))
	state.MSSClamp = stringValueOrNull(setting.MSSClamp)
	state.OffloadAccounting = types.BoolValue(derefBool(setting.OffloadAccounting))
	state.OffloadL2Blocking = types.BoolValue(derefBool(setting.OffloadL2Blocking))
	state.OffloadSch = types.BoolValue(derefBool(setting.OffloadSch))
	state.PPTPModule = types.BoolValue(derefBool(setting.PPTPModule))
	state.ReceiveRedirects = types.BoolValue(derefBool(setting.ReceiveRedirects))
	state.SendRedirects = types.BoolValue(derefBool(setting.SendRedirects))
	state.SIPModule = types.BoolValue(derefBool(setting.SIPModule))
	state.SynCookies = types.BoolValue(derefBool(setting.SynCookies))
	state.TFTPModule = types.BoolValue(derefBool(setting.TFTPModule))
	state.UPnPEnabled = types.BoolValue(derefBool(setting.UPnPEnabled))
	state.UPnPNATPMPEnabled = types.BoolValue(derefBool(setting.UPnPNATPMPEnabled))
	state.UPnPSecureMode = types.BoolValue(derefBool(setting.UPnPSecureMode))

	if setting.MSSClampMSS != nil {
		state.MSSClampMSS = types.Int64Value(int64(*setting.MSSClampMSS))
	} else {
		state.MSSClampMSS = types.Int64Null()
	}

	return diags
}
