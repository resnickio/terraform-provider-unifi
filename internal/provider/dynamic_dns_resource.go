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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &DynamicDNSResource{}
	_ resource.ResourceWithImportState = &DynamicDNSResource{}
)

type DynamicDNSResource struct {
	client *AutoLoginClient
}

type DynamicDNSResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	SiteID    types.String   `tfsdk:"site_id"`
	Service   types.String   `tfsdk:"service"`
	HostName  types.String   `tfsdk:"hostname"`
	Login     types.String   `tfsdk:"login"`
	Password  types.String   `tfsdk:"password"`
	Server    types.String   `tfsdk:"server"`
	Interface types.String   `tfsdk:"interface"`
	Options   types.String   `tfsdk:"options"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func NewDynamicDNSResource() resource.Resource {
	return &DynamicDNSResource{}
}

func (r *DynamicDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_dns"
}

func (r *DynamicDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi dynamic DNS configuration.",
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
				Description: "The unique identifier of the dynamic DNS configuration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the dynamic DNS is configured.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service": schema.StringAttribute{
				Description: "The dynamic DNS service provider. Valid values: afraid, changeip, cloudflare, dnspark, dslreports, dyndns, easydns, namecheap, noip, sitelutions, zoneedit, custom.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("afraid", "changeip", "cloudflare", "dnspark", "dslreports", "dyndns", "easydns", "namecheap", "noip", "sitelutions", "zoneedit", "custom"),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname to update with the dynamic DNS service.",
				Required:    true,
			},
			"login": schema.StringAttribute{
				Description: "The login/username for the dynamic DNS service.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password or API token for the dynamic DNS service. Note: This value is write-only and cannot be read back from the controller.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server": schema.StringAttribute{
				Description: "The server address for the dynamic DNS service (primarily used with 'custom' service).",
				Optional:    true,
			},
			"interface": schema.StringAttribute{
				Description: "The WAN interface to monitor for IP changes. Valid values: wan, wan2. Defaults to wan.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("wan"),
				Validators: []validator.String{
					stringvalidator.OneOf("wan", "wan2"),
				},
			},
			"options": schema.StringAttribute{
				Description: "Additional options for the dynamic DNS service.",
				Optional:    true,
			},
		},
	}
}

func (r *DynamicDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DynamicDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DynamicDNSResourceModel

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

	dns := r.planToSDK(&plan)

	created, err := r.client.CreateDynamicDNS(ctx, dns)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "dynamic DNS")
		return
	}

	// Save password from plan (API won't return it)
	originalPassword := plan.Password

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore password if it was set
	if !originalPassword.IsNull() {
		plan.Password = originalPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DynamicDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DynamicDNSResourceModel

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

	dns, err := r.client.GetDynamicDNS(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "dynamic DNS")
		return
	}

	// Save prior state for password preservation
	priorState := state

	resp.Diagnostics.Append(r.sdkToState(ctx, dns, &state, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DynamicDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DynamicDNSResourceModel

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

	var state DynamicDNSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dns := r.planToSDK(&plan)
	dns.ID = state.ID.ValueString()
	dns.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateDynamicDNS(ctx, state.ID.ValueString(), dns)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "dynamic DNS")
		return
	}

	// Save password from plan (API won't return it)
	originalPassword := plan.Password

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore password if it was set
	if !originalPassword.IsNull() {
		plan.Password = originalPassword
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DynamicDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DynamicDNSResourceModel

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

	err := r.client.DeleteDynamicDNS(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "dynamic DNS")
		return
	}
}

func (r *DynamicDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DynamicDNSResource) planToSDK(plan *DynamicDNSResourceModel) *unifi.DynamicDNS {
	dns := &unifi.DynamicDNS{
		Service:   plan.Service.ValueString(),
		HostName:  plan.HostName.ValueString(),
		Interface: plan.Interface.ValueString(),
	}

	if !plan.Login.IsNull() && !plan.Login.IsUnknown() {
		dns.Login = plan.Login.ValueString()
	}

	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		dns.XPassword = plan.Password.ValueString()
	}

	if !plan.Server.IsNull() && !plan.Server.IsUnknown() {
		dns.Server = plan.Server.ValueString()
	}

	if !plan.Options.IsNull() && !plan.Options.IsUnknown() {
		dns.Options = plan.Options.ValueString()
	}

	return dns
}

func (r *DynamicDNSResource) sdkToState(ctx context.Context, dns *unifi.DynamicDNS, state *DynamicDNSResourceModel, priorState *DynamicDNSResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(dns.ID)
	state.SiteID = stringValueOrNull(dns.SiteID)
	state.Service = types.StringValue(dns.Service)
	state.HostName = types.StringValue(dns.HostName)
	state.Interface = stringValueOrNull(dns.Interface)

	if dns.Login != "" {
		state.Login = types.StringValue(dns.Login)
	} else {
		state.Login = types.StringNull()
	}

	if dns.Server != "" {
		state.Server = types.StringValue(dns.Server)
	} else {
		state.Server = types.StringNull()
	}

	if dns.Options != "" {
		state.Options = types.StringValue(dns.Options)
	} else {
		state.Options = types.StringNull()
	}

	// Password is write-only - preserve from prior state to prevent drift
	if priorState != nil && !priorState.Password.IsNull() {
		state.Password = priorState.Password
	}

	return diags
}
