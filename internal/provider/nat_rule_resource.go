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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &NatRuleResource{}
	_ resource.ResourceWithImportState = &NatRuleResource{}
)

type NatRuleResource struct {
	client *AutoLoginClient
}

type NatRuleResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Enabled        types.Bool     `tfsdk:"enabled"`
	Type           types.String   `tfsdk:"type"`
	Description    types.String   `tfsdk:"description"`
	Protocol       types.String   `tfsdk:"protocol"`
	SourceAddress  types.String   `tfsdk:"source_address"`
	SourcePort     types.String   `tfsdk:"source_port"`
	DestAddress    types.String   `tfsdk:"dest_address"`
	DestPort       types.String   `tfsdk:"dest_port"`
	TranslatedIP   types.String   `tfsdk:"translated_ip"`
	TranslatedPort types.String   `tfsdk:"translated_port"`
	Logging        types.Bool     `tfsdk:"logging"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func NewNatRuleResource() resource.Resource {
	return &NatRuleResource{}
}

func (r *NatRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_rule"
}

func (r *NatRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi NAT rule.",
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
				Description: "The unique identifier of the NAT rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the NAT rule is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "The NAT rule type. Valid values: MASQUERADE, DNAT, SNAT.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("MASQUERADE", "DNAT", "SNAT"),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for the NAT rule.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol for the NAT rule. Valid values: all, tcp, udp, tcp_udp. Defaults to all.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "tcp", "udp", "tcp_udp"),
				},
			},
			"source_address": schema.StringAttribute{
				Description: "The source IP address or CIDR block.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_port": schema.StringAttribute{
				Description: "The source port or port range (e.g., '80' or '8000-9000').",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dest_address": schema.StringAttribute{
				Description: "The destination IP address or CIDR block.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dest_port": schema.StringAttribute{
				Description: "The destination port or port range.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"translated_ip": schema.StringAttribute{
				Description: "The IP address to translate to (for DNAT/SNAT).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"translated_port": schema.StringAttribute{
				Description: "The port to translate to (for DNAT/SNAT).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"logging": schema.BoolAttribute{
				Description: "Whether to log traffic matching this rule. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *NatRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NatRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NatRuleResourceModel

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

	rule := r.planToSDK(&plan)

	created, err := r.client.CreateNatRule(ctx, rule)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "NAT rule")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NatRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NatRuleResourceModel

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

	rule, err := r.client.GetNatRule(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "NAT rule")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, rule, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NatRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NatRuleResourceModel

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

	var state NatRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := r.planToSDK(&plan)
	rule.ID = state.ID.ValueString()

	updated, err := r.client.UpdateNatRule(ctx, state.ID.ValueString(), rule)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "NAT rule")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NatRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NatRuleResourceModel

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

	err := r.client.DeleteNatRule(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "NAT rule")
		return
	}
}

func (r *NatRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NatRuleResource) planToSDK(plan *NatRuleResourceModel) *unifi.NatRule {
	rule := &unifi.NatRule{
		Enabled:  boolPtr(plan.Enabled.ValueBool()),
		Type:     plan.Type.ValueString(),
		Protocol: plan.Protocol.ValueString(),
		Logging:  boolPtr(plan.Logging.ValueBool()),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		rule.Description = plan.Description.ValueString()
	}

	if !plan.SourceAddress.IsNull() && !plan.SourceAddress.IsUnknown() {
		rule.SourceAddress = plan.SourceAddress.ValueString()
	}

	if !plan.SourcePort.IsNull() && !plan.SourcePort.IsUnknown() {
		rule.SourcePort = plan.SourcePort.ValueString()
	}

	if !plan.DestAddress.IsNull() && !plan.DestAddress.IsUnknown() {
		rule.DestAddress = plan.DestAddress.ValueString()
	}

	if !plan.DestPort.IsNull() && !plan.DestPort.IsUnknown() {
		rule.DestPort = plan.DestPort.ValueString()
	}

	if !plan.TranslatedIP.IsNull() && !plan.TranslatedIP.IsUnknown() {
		rule.TranslatedIP = plan.TranslatedIP.ValueString()
	}

	if !plan.TranslatedPort.IsNull() && !plan.TranslatedPort.IsUnknown() {
		rule.TranslatedPort = plan.TranslatedPort.ValueString()
	}

	return rule
}

func (r *NatRuleResource) sdkToState(ctx context.Context, rule *unifi.NatRule, state *NatRuleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(rule.ID)
	state.Enabled = types.BoolValue(derefBool(rule.Enabled))
	state.Type = types.StringValue(rule.Type)
	state.Protocol = stringValueOrNull(rule.Protocol)
	state.Logging = types.BoolValue(derefBool(rule.Logging))

	if rule.Description != "" {
		state.Description = types.StringValue(rule.Description)
	} else {
		state.Description = types.StringNull()
	}

	if rule.SourceAddress != "" {
		state.SourceAddress = types.StringValue(rule.SourceAddress)
	} else {
		state.SourceAddress = types.StringNull()
	}

	if rule.SourcePort != "" {
		state.SourcePort = types.StringValue(rule.SourcePort)
	} else {
		state.SourcePort = types.StringNull()
	}

	if rule.DestAddress != "" {
		state.DestAddress = types.StringValue(rule.DestAddress)
	} else {
		state.DestAddress = types.StringNull()
	}

	if rule.DestPort != "" {
		state.DestPort = types.StringValue(rule.DestPort)
	} else {
		state.DestPort = types.StringNull()
	}

	if rule.TranslatedIP != "" {
		state.TranslatedIP = types.StringValue(rule.TranslatedIP)
	} else {
		state.TranslatedIP = types.StringNull()
	}

	if rule.TranslatedPort != "" {
		state.TranslatedPort = types.StringValue(rule.TranslatedPort)
	} else {
		state.TranslatedPort = types.StringNull()
	}

	return diags
}
