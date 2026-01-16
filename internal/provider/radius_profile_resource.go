package provider

import (
	"context"
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
	_ resource.Resource                = &RADIUSProfileResource{}
	_ resource.ResourceWithImportState = &RADIUSProfileResource{}
)

type RADIUSProfileResource struct {
	client *AutoLoginClient
}

type RADIUSProfileResourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	SiteID                types.String   `tfsdk:"site_id"`
	Name                  types.String   `tfsdk:"name"`
	UseUsgAuthServer      types.Bool     `tfsdk:"use_usg_auth_server"`
	UseUsgAcctServer      types.Bool     `tfsdk:"use_usg_acct_server"`
	VlanEnabled           types.Bool     `tfsdk:"vlan_enabled"`
	VlanWlanMode          types.String   `tfsdk:"vlan_wlan_mode"`
	InterimUpdateEnabled  types.Bool     `tfsdk:"interim_update_enabled"`
	InterimUpdateInterval types.Int64    `tfsdk:"interim_update_interval"`
	AuthServers           types.List     `tfsdk:"auth_server"`
	AcctServers           types.List     `tfsdk:"acct_server"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

type RADIUSServerModel struct {
	IP     types.String `tfsdk:"ip"`
	Port   types.Int64  `tfsdk:"port"`
	Secret types.String `tfsdk:"secret"`
}

var radiusServerAttrTypes = map[string]attr.Type{
	"ip":     types.StringType,
	"port":   types.Int64Type,
	"secret": types.StringType,
}

func NewRADIUSProfileResource() resource.Resource {
	return &RADIUSProfileResource{}
}

func (r *RADIUSProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radius_profile"
}

func (r *RADIUSProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	serverSchema := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				Description: "The IP address of the RADIUS server.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port of the RADIUS server.",
				Optional:    true,
			},
			"secret": schema.StringAttribute{
				Description: "The shared secret for the RADIUS server.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a UniFi RADIUS profile for authentication.",
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the RADIUS profile.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the RADIUS profile exists.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the RADIUS profile.",
				Required:    true,
			},
			"use_usg_auth_server": schema.BoolAttribute{
				Description: "Use the USG/UDM as the authentication server. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"use_usg_acct_server": schema.BoolAttribute{
				Description: "Use the USG/UDM as the accounting server. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"vlan_enabled": schema.BoolAttribute{
				Description: "Enable VLAN assignment via RADIUS. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"vlan_wlan_mode": schema.StringAttribute{
				Description: "VLAN WLAN mode. Valid values: disabled, optional, required.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "optional", "required"),
				},
			},
			"interim_update_enabled": schema.BoolAttribute{
				Description: "Enable interim accounting updates. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"interim_update_interval": schema.Int64Attribute{
				Description: "Interval in seconds between interim accounting updates.",
				Optional:    true,
			},
			"auth_server": schema.ListNestedAttribute{
				Description: "List of RADIUS authentication servers.",
				Optional:    true,
				NestedObject: serverSchema,
			},
			"acct_server": schema.ListNestedAttribute{
				Description: "List of RADIUS accounting servers.",
				Optional:    true,
				NestedObject: serverSchema,
			},
		},
	}
}

func (r *RADIUSProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RADIUSProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RADIUSProfileResourceModel

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

	profile := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateRADIUSProfile(ctx, profile)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "RADIUS profile")
		return
	}

	priorSecrets := r.extractSecrets(ctx, &plan)

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan, priorSecrets)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RADIUSProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RADIUSProfileResourceModel

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

	profile, err := r.client.GetRADIUSProfile(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "RADIUS profile")
		return
	}

	priorSecrets := r.extractSecrets(ctx, &state)

	resp.Diagnostics.Append(r.sdkToState(ctx, profile, &state, priorSecrets)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RADIUSProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RADIUSProfileResourceModel

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

	var state RADIUSProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	profile.ID = state.ID.ValueString()
	profile.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateRADIUSProfile(ctx, state.ID.ValueString(), profile)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "RADIUS profile")
		return
	}

	priorSecrets := r.extractSecrets(ctx, &plan)

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan, priorSecrets)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RADIUSProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RADIUSProfileResourceModel

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

	err := r.client.DeleteRADIUSProfile(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "RADIUS profile")
		return
	}
}

func (r *RADIUSProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type serverSecrets struct {
	AuthSecrets []string
	AcctSecrets []string
}

func (r *RADIUSProfileResource) extractSecrets(ctx context.Context, model *RADIUSProfileResourceModel) serverSecrets {
	secrets := serverSecrets{}

	if !model.AuthServers.IsNull() && !model.AuthServers.IsUnknown() {
		var servers []RADIUSServerModel
		model.AuthServers.ElementsAs(ctx, &servers, false)
		for _, s := range servers {
			if !s.Secret.IsNull() && !s.Secret.IsUnknown() {
				secrets.AuthSecrets = append(secrets.AuthSecrets, s.Secret.ValueString())
			} else {
				secrets.AuthSecrets = append(secrets.AuthSecrets, "")
			}
		}
	}

	if !model.AcctServers.IsNull() && !model.AcctServers.IsUnknown() {
		var servers []RADIUSServerModel
		model.AcctServers.ElementsAs(ctx, &servers, false)
		for _, s := range servers {
			if !s.Secret.IsNull() && !s.Secret.IsUnknown() {
				secrets.AcctSecrets = append(secrets.AcctSecrets, s.Secret.ValueString())
			} else {
				secrets.AcctSecrets = append(secrets.AcctSecrets, "")
			}
		}
	}

	return secrets
}

func (r *RADIUSProfileResource) planToSDK(ctx context.Context, plan *RADIUSProfileResourceModel, diags *diag.Diagnostics) *unifi.RADIUSProfile {
	profile := &unifi.RADIUSProfile{
		Name:                 plan.Name.ValueString(),
		UseUsgAuthServer:     boolPtr(plan.UseUsgAuthServer.ValueBool()),
		UseUsgAcctServer:     boolPtr(plan.UseUsgAcctServer.ValueBool()),
		VlanEnabled:          boolPtr(plan.VlanEnabled.ValueBool()),
		InterimUpdateEnabled: boolPtr(plan.InterimUpdateEnabled.ValueBool()),
	}

	if !plan.VlanWlanMode.IsNull() && !plan.VlanWlanMode.IsUnknown() {
		profile.VlanWlanMode = plan.VlanWlanMode.ValueString()
	}

	if !plan.InterimUpdateInterval.IsNull() && !plan.InterimUpdateInterval.IsUnknown() {
		interval := int(plan.InterimUpdateInterval.ValueInt64())
		profile.InterimUpdateInterval = &interval
	}

	if !plan.AuthServers.IsNull() && !plan.AuthServers.IsUnknown() {
		var servers []RADIUSServerModel
		diags.Append(plan.AuthServers.ElementsAs(ctx, &servers, false)...)
		for _, s := range servers {
			server := unifi.RADIUSServer{
				IP: s.IP.ValueString(),
			}
			if !s.Port.IsNull() && !s.Port.IsUnknown() {
				port := int(s.Port.ValueInt64())
				server.Port = &port
			}
			if !s.Secret.IsNull() && !s.Secret.IsUnknown() {
				server.XSecret = s.Secret.ValueString()
			}
			profile.AuthServers = append(profile.AuthServers, server)
		}
	}

	if !plan.AcctServers.IsNull() && !plan.AcctServers.IsUnknown() {
		var servers []RADIUSServerModel
		diags.Append(plan.AcctServers.ElementsAs(ctx, &servers, false)...)
		for _, s := range servers {
			server := unifi.RADIUSServer{
				IP: s.IP.ValueString(),
			}
			if !s.Port.IsNull() && !s.Port.IsUnknown() {
				port := int(s.Port.ValueInt64())
				server.Port = &port
			}
			if !s.Secret.IsNull() && !s.Secret.IsUnknown() {
				server.XSecret = s.Secret.ValueString()
			}
			profile.AcctServers = append(profile.AcctServers, server)
		}
	}

	return profile
}

func (r *RADIUSProfileResource) sdkToState(ctx context.Context, profile *unifi.RADIUSProfile, state *RADIUSProfileResourceModel, secrets serverSecrets) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(profile.ID)
	state.SiteID = stringValueOrNull(profile.SiteID)
	state.Name = types.StringValue(profile.Name)
	state.UseUsgAuthServer = types.BoolValue(derefBool(profile.UseUsgAuthServer))
	state.UseUsgAcctServer = types.BoolValue(derefBool(profile.UseUsgAcctServer))
	state.VlanEnabled = types.BoolValue(derefBool(profile.VlanEnabled))
	state.InterimUpdateEnabled = types.BoolValue(derefBool(profile.InterimUpdateEnabled))

	if profile.VlanWlanMode != "" {
		state.VlanWlanMode = types.StringValue(profile.VlanWlanMode)
	} else {
		state.VlanWlanMode = types.StringNull()
	}

	if profile.InterimUpdateInterval != nil {
		state.InterimUpdateInterval = types.Int64Value(int64(*profile.InterimUpdateInterval))
	} else {
		state.InterimUpdateInterval = types.Int64Null()
	}

	if len(profile.AuthServers) > 0 {
		var elements []attr.Value
		for i, s := range profile.AuthServers {
			var port types.Int64
			if s.Port != nil {
				port = types.Int64Value(int64(*s.Port))
			} else {
				port = types.Int64Null()
			}

			var secret types.String
			if i < len(secrets.AuthSecrets) && secrets.AuthSecrets[i] != "" {
				secret = types.StringValue(secrets.AuthSecrets[i])
			} else {
				secret = types.StringNull()
			}

			attrs := map[string]attr.Value{
				"ip":     types.StringValue(s.IP),
				"port":   port,
				"secret": secret,
			}
			obj, d := types.ObjectValue(radiusServerAttrTypes, attrs)
			diags.Append(d...)
			elements = append(elements, obj)
		}
		list, d := types.ListValue(types.ObjectType{AttrTypes: radiusServerAttrTypes}, elements)
		diags.Append(d...)
		state.AuthServers = list
	} else {
		state.AuthServers = types.ListNull(types.ObjectType{AttrTypes: radiusServerAttrTypes})
	}

	if len(profile.AcctServers) > 0 {
		var elements []attr.Value
		for i, s := range profile.AcctServers {
			var port types.Int64
			if s.Port != nil {
				port = types.Int64Value(int64(*s.Port))
			} else {
				port = types.Int64Null()
			}

			var secret types.String
			if i < len(secrets.AcctSecrets) && secrets.AcctSecrets[i] != "" {
				secret = types.StringValue(secrets.AcctSecrets[i])
			} else {
				secret = types.StringNull()
			}

			attrs := map[string]attr.Value{
				"ip":     types.StringValue(s.IP),
				"port":   port,
				"secret": secret,
			}
			obj, d := types.ObjectValue(radiusServerAttrTypes, attrs)
			diags.Append(d...)
			elements = append(elements, obj)
		}
		list, d := types.ListValue(types.ObjectType{AttrTypes: radiusServerAttrTypes}, elements)
		diags.Append(d...)
		state.AcctServers = list
	} else {
		state.AcctServers = types.ListNull(types.ObjectType{AttrTypes: radiusServerAttrTypes})
	}

	return diags
}
