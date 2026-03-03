package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &AccountResource{}
	_ resource.ResourceWithImportState = &AccountResource{}
)

type AccountResource struct {
	client *AutoLoginClient
}

type AccountResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	SiteID           types.String   `tfsdk:"site_id"`
	Name             types.String   `tfsdk:"name"`
	XPassword        types.String   `tfsdk:"x_password"`
	TunnelConfigType types.String   `tfsdk:"tunnel_config_type"`
	TunnelMediumType types.Int64    `tfsdk:"tunnel_medium_type"`
	TunnelType       types.Int64    `tfsdk:"tunnel_type"`
	VLAN             types.Int64    `tfsdk:"vlan"`
	NetworkConfID    types.String   `tfsdk:"network_id"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func NewAccountResource() resource.Resource {
	return &AccountResource{}
}

func (r *AccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r *AccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi RADIUS account for 802.1X or VPN authentication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the RADIUS account.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The username for the RADIUS account.",
				Required:    true,
			},
			"x_password": schema.StringAttribute{
				Description: "The password for the RADIUS account (write-only).",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tunnel_config_type": schema.StringAttribute{
				Description: "Tunnel configuration type. Valid values: '802.1x', 'vpn', 'custom'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("802.1x", "vpn", "custom"),
				},
			},
			"tunnel_medium_type": schema.Int64Attribute{
				Description: "Tunnel medium type (1-15).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 15),
				},
			},
			"tunnel_type": schema.Int64Attribute{
				Description: "Tunnel type (1-13).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 13),
				},
			},
			"vlan": schema.Int64Attribute{
				Description: "VLAN ID (2-4009).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(2, 4009),
				},
			},
			"network_id": schema.StringAttribute{
				Description: "The network configuration ID.",
				Optional:    true,
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

func (r *AccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccountResourceModel
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

	account := r.planToSDK(&plan)

	savedPassword := plan.XPassword

	created, err := r.client.CreateRADIUSAccount(ctx, account)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "RADIUS account")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XPassword = savedPassword
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountResourceModel
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

	savedPassword := state.XPassword

	account, err := r.client.GetRADIUSAccount(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "RADIUS account")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(account, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.XPassword.IsNull() || state.XPassword.ValueString() == "" {
		state.XPassword = savedPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AccountResourceModel
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

	var state AccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account := r.planToSDK(&plan)
	account.ID = state.ID.ValueString()

	savedPassword := plan.XPassword

	updated, err := r.client.UpdateRADIUSAccount(ctx, state.ID.ValueString(), account)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "RADIUS account")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.XPassword = savedPassword
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountResourceModel
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

	err := r.client.DeleteRADIUSAccount(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "RADIUS account")
		return
	}
}

func (r *AccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AccountResource) planToSDK(plan *AccountResourceModel) *unifi.RADIUSAccount {
	account := &unifi.RADIUSAccount{
		Name: plan.Name.ValueString(),
	}

	if !plan.XPassword.IsNull() && !plan.XPassword.IsUnknown() {
		account.XPassword = plan.XPassword.ValueString()
	}
	if !plan.TunnelConfigType.IsNull() {
		account.TunnelConfigType = plan.TunnelConfigType.ValueString()
	}
	if !plan.TunnelMediumType.IsNull() {
		v := int(plan.TunnelMediumType.ValueInt64())
		account.TunnelMediumType = &v
	}
	if !plan.TunnelType.IsNull() {
		v := int(plan.TunnelType.ValueInt64())
		account.TunnelType = &v
	}
	if !plan.VLAN.IsNull() {
		v := int(plan.VLAN.ValueInt64())
		account.VLAN = &v
	}
	if !plan.NetworkConfID.IsNull() {
		account.NetworkConfID = plan.NetworkConfID.ValueString()
	}

	return account
}

func (r *AccountResource) sdkToState(account *unifi.RADIUSAccount, state *AccountResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(account.ID)
	state.SiteID = types.StringValue(account.SiteID)
	state.Name = types.StringValue(account.Name)
	state.TunnelConfigType = stringValueOrNull(account.TunnelConfigType)
	state.NetworkConfID = stringValueOrNull(account.NetworkConfID)

	if account.XPassword != "" {
		state.XPassword = types.StringValue(account.XPassword)
	}

	if account.TunnelMediumType != nil {
		state.TunnelMediumType = types.Int64Value(int64(*account.TunnelMediumType))
	} else {
		state.TunnelMediumType = types.Int64Null()
	}

	if account.TunnelType != nil {
		state.TunnelType = types.Int64Value(int64(*account.TunnelType))
	} else {
		state.TunnelType = types.Int64Null()
	}

	if account.VLAN != nil {
		state.VLAN = types.Int64Value(int64(*account.VLAN))
	} else {
		state.VLAN = types.Int64Null()
	}

	return diags
}
