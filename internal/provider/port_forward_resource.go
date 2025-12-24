package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

const (
	defaultPfwdInterface = "wan"
)

var (
	_ resource.Resource                = &PortForwardResource{}
	_ resource.ResourceWithImportState = &PortForwardResource{}
)

type PortForwardResource struct {
	client *AutoLoginClient
}

type PortForwardResourceModel struct {
	ID            types.String `tfsdk:"id"`
	SiteID        types.String `tfsdk:"site_id"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Protocol      types.String `tfsdk:"protocol"`
	DstPort       types.String `tfsdk:"dst_port"`
	FwdPort       types.String `tfsdk:"fwd_port"`
	FwdIP         types.String `tfsdk:"fwd_ip"`
	Src           types.String `tfsdk:"src"`
	PfwdInterface types.String `tfsdk:"pfwd_interface"`
	Log           types.Bool   `tfsdk:"log"`
}

func NewPortForwardResource() resource.Resource {
	return &PortForwardResource{}
}

func (r *PortForwardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_forward"
}

func (r *PortForwardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi port forwarding rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the port forward rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the port forward rule is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the port forward rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the port forward rule is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol for port forwarding. Valid values: 'tcp', 'udp', 'tcp_udp'.",
				Required:    true,
			},
			"dst_port": schema.StringAttribute{
				Description: "The destination port to forward from (external port).",
				Required:    true,
			},
			"fwd_port": schema.StringAttribute{
				Description: "The port to forward to on the destination host.",
				Required:    true,
			},
			"fwd_ip": schema.StringAttribute{
				Description: "The IP address to forward traffic to.",
				Required:    true,
			},
			"src": schema.StringAttribute{
				Description: "Restrict forwarding to traffic from this source IP/CIDR. Leave empty for any source.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"pfwd_interface": schema.StringAttribute{
				Description: "The WAN interface for the port forward. Valid values: 'wan', 'wan2', 'both'. Defaults to 'wan'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(defaultPfwdInterface),
			},
			"log": schema.BoolAttribute{
				Description: "Whether to log forwarded traffic. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *PortForwardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PortForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PortForwardResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert plan to SDK struct
	pf := r.planToSDK(&plan)

	// Create the port forward
	created, err := r.client.CreatePortForward(ctx, pf)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "port forward")
		return
	}

	// Update state with response
	resp.Diagnostics.Append(r.sdkToState(created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PortForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PortForwardResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the port forward
	pf, err := r.client.GetPortForward(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "port forward")
		return
	}

	// Update state with response
	resp.Diagnostics.Append(r.sdkToState(pf, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PortForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PortForwardResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PortForwardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert plan to SDK struct
	pf := r.planToSDK(&plan)

	// Preserve ID and SiteID from state
	pf.ID = state.ID.ValueString()
	pf.SiteID = state.SiteID.ValueString()

	// Update the port forward
	updated, err := r.client.UpdatePortForward(ctx, state.ID.ValueString(), pf)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "port forward")
		return
	}

	// Update state with response
	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PortForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PortForwardResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the port forward
	err := r.client.DeletePortForward(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "port forward")
		return
	}
}

func (r *PortForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// planToSDK converts the Terraform plan to an SDK PortForward struct.
func (r *PortForwardResource) planToSDK(plan *PortForwardResourceModel) *unifi.PortForward {
	pf := &unifi.PortForward{
		Name:          plan.Name.ValueString(),
		Enabled:       boolPtr(plan.Enabled.ValueBool()),
		Proto:         plan.Protocol.ValueString(),
		DstPort:       plan.DstPort.ValueString(),
		FwdPort:       plan.FwdPort.ValueString(),
		Fwd:           plan.FwdIP.ValueString(),
		Src:           plan.Src.ValueString(),
		PfwdInterface: plan.PfwdInterface.ValueString(),
		Log:           boolPtr(plan.Log.ValueBool()),
	}

	return pf
}

// sdkToState updates the Terraform state from an SDK PortForward struct.
func (r *PortForwardResource) sdkToState(pf *unifi.PortForward, state *PortForwardResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(pf.ID)
	state.SiteID = types.StringValue(pf.SiteID)
	state.Name = types.StringValue(pf.Name)
	state.Enabled = types.BoolValue(derefBool(pf.Enabled))
	state.Protocol = types.StringValue(pf.Proto)
	state.DstPort = types.StringValue(pf.DstPort)
	state.FwdPort = types.StringValue(pf.FwdPort)
	state.FwdIP = types.StringValue(pf.Fwd)
	state.Src = types.StringValue(pf.Src)

	if pf.PfwdInterface != "" {
		state.PfwdInterface = types.StringValue(pf.PfwdInterface)
	} else {
		state.PfwdInterface = types.StringValue(defaultPfwdInterface)
	}

	state.Log = types.BoolValue(derefBool(pf.Log))

	return diags
}
