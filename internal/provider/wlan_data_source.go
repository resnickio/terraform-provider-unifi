package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ datasource.DataSource = &WLANDataSource{}

type WLANDataSource struct {
	client *AutoLoginClient
}

type WLANDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	SiteID           types.String `tfsdk:"site_id"`
	Name             types.String `tfsdk:"name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Security         types.String `tfsdk:"security"`
	WPAMode          types.String `tfsdk:"wpa_mode"`
	WPAEnc           types.String `tfsdk:"wpa_enc"`
	NetworkID        types.String `tfsdk:"network_id"`
	UserGroupID      types.String `tfsdk:"user_group_id"`
	APGroupIDs       types.Set    `tfsdk:"ap_group_ids"`
	IsGuest          types.Bool   `tfsdk:"is_guest"`
	HideSsid         types.Bool   `tfsdk:"hide_ssid"`
	WLANBand         types.String `tfsdk:"wlan_band"`
	WLANBands        types.Set    `tfsdk:"wlan_bands"`
	Vlan             types.Int64  `tfsdk:"vlan"`
	VlanEnabled      types.Bool   `tfsdk:"vlan_enabled"`
	MacFilterEnabled types.Bool   `tfsdk:"mac_filter_enabled"`
	MacFilterList    types.Set    `tfsdk:"mac_filter_list"`
	MacFilterPolicy  types.String `tfsdk:"mac_filter_policy"`
	ScheduleEnabled  types.Bool   `tfsdk:"schedule_enabled"`
	L2Isolation      types.Bool   `tfsdk:"l2_isolation"`
	FastRoaming      types.Bool   `tfsdk:"fast_roaming_enabled"`
	ProxyArp         types.Bool   `tfsdk:"proxy_arp"`
	BssTransition    types.Bool   `tfsdk:"bss_transition"`
	Uapsd            types.Bool   `tfsdk:"uapsd_enabled"`
	PmfMode          types.String `tfsdk:"pmf_mode"`
	WPA3Support      types.Bool   `tfsdk:"wpa3_support"`
	WPA3Transition   types.Bool   `tfsdk:"wpa3_transition"`
}

func NewWLANDataSource() datasource.DataSource {
	return &WLANDataSource{}
}

func (d *WLANDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wlan"
}

func (d *WLANDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing UniFi wireless network (WLAN/SSID). Lookup by either id or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the WLAN. Specify either id or name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The SSID name of the wireless network. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the WLAN exists.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the WLAN is enabled.",
				Computed:    true,
			},
			"security": schema.StringAttribute{
				Description: "The security mode (open, wep, wpapsk, wpaeap).",
				Computed:    true,
			},
			"wpa_mode": schema.StringAttribute{
				Description: "The WPA mode (auto, wpa1, wpa2).",
				Computed:    true,
			},
			"wpa_enc": schema.StringAttribute{
				Description: "The WPA encryption type.",
				Computed:    true,
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID (VLAN) assigned to this WLAN.",
				Computed:    true,
			},
			"user_group_id": schema.StringAttribute{
				Description: "The user group ID for bandwidth limiting.",
				Computed:    true,
			},
			"ap_group_ids": schema.SetAttribute{
				Description: "Set of AP group IDs this WLAN is broadcast on.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_guest": schema.BoolAttribute{
				Description: "Whether this is a guest network.",
				Computed:    true,
			},
			"hide_ssid": schema.BoolAttribute{
				Description: "Whether the SSID is hidden.",
				Computed:    true,
			},
			"wlan_band": schema.StringAttribute{
				Description: "The wireless band (2g, 5g, both).",
				Computed:    true,
			},
			"wlan_bands": schema.SetAttribute{
				Description: "Set of wireless bands enabled.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"vlan": schema.Int64Attribute{
				Description: "The VLAN ID for this WLAN.",
				Computed:    true,
			},
			"vlan_enabled": schema.BoolAttribute{
				Description: "Whether VLAN tagging is enabled.",
				Computed:    true,
			},
			"mac_filter_enabled": schema.BoolAttribute{
				Description: "Whether MAC filtering is enabled.",
				Computed:    true,
			},
			"mac_filter_list": schema.SetAttribute{
				Description: "Set of MAC addresses for filtering.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"mac_filter_policy": schema.StringAttribute{
				Description: "MAC filter policy (allow, deny).",
				Computed:    true,
			},
			"schedule_enabled": schema.BoolAttribute{
				Description: "Whether scheduling is enabled.",
				Computed:    true,
			},
			"l2_isolation": schema.BoolAttribute{
				Description: "Whether L2 isolation is enabled.",
				Computed:    true,
			},
			"fast_roaming_enabled": schema.BoolAttribute{
				Description: "Whether fast roaming (802.11r) is enabled.",
				Computed:    true,
			},
			"proxy_arp": schema.BoolAttribute{
				Description: "Whether proxy ARP is enabled.",
				Computed:    true,
			},
			"bss_transition": schema.BoolAttribute{
				Description: "Whether BSS transition is enabled.",
				Computed:    true,
			},
			"uapsd_enabled": schema.BoolAttribute{
				Description: "Whether U-APSD (WMM Power Save) is enabled.",
				Computed:    true,
			},
			"pmf_mode": schema.StringAttribute{
				Description: "Protected Management Frames mode.",
				Computed:    true,
			},
			"wpa3_support": schema.BoolAttribute{
				Description: "Whether WPA3 support is enabled.",
				Computed:    true,
			},
			"wpa3_transition": schema.BoolAttribute{
				Description: "Whether WPA3 transition mode is enabled.",
				Computed:    true,
			},
		},
	}
}

func (d *WLANDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WLANDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WLANDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a WLAN.",
		)
		return
	}

	var wlan *unifi.WLANConf
	var err error

	if hasID {
		wlan, err = d.client.GetWLAN(ctx, config.ID.ValueString())
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "read", "WLAN")
			return
		}
	} else {
		wlans, err := d.client.ListWLANs(ctx)
		if err != nil {
			handleSDKError(&resp.Diagnostics, err, "list", "WLANs")
			return
		}

		searchName := config.Name.ValueString()
		for i := range wlans {
			if wlans[i].Name == searchName {
				wlan = &wlans[i]
				break
			}
		}

		if wlan == nil {
			resp.Diagnostics.AddError(
				"WLAN Not Found",
				fmt.Sprintf("No WLAN found with name '%s'.", searchName),
			)
			return
		}
	}

	resp.Diagnostics.Append(d.sdkToState(ctx, wlan, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *WLANDataSource) sdkToState(ctx context.Context, wlan *unifi.WLANConf, state *WLANDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(wlan.ID)
	state.SiteID = types.StringValue(wlan.SiteID)
	state.Name = types.StringValue(wlan.Name)
	state.Enabled = types.BoolValue(derefBool(wlan.Enabled))
	state.Security = types.StringValue(wlan.Security)
	state.WPAMode = types.StringValue(wlan.WPAMode)
	state.WPAEnc = types.StringValue(wlan.WPAEnc)

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
