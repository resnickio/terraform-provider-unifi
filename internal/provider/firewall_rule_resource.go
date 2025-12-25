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
	_ resource.Resource                = &FirewallRuleResource{}
	_ resource.ResourceWithImportState = &FirewallRuleResource{}
)

type FirewallRuleResource struct {
	client *AutoLoginClient
}

type FirewallRuleResourceModel struct {
	ID                  types.String   `tfsdk:"id"`
	SiteID              types.String   `tfsdk:"site_id"`
	Name                types.String   `tfsdk:"name"`
	Ruleset             types.String   `tfsdk:"ruleset"`
	Action              types.String   `tfsdk:"action"`
	RuleIndex           types.Int64    `tfsdk:"rule_index"`
	Enabled             types.Bool     `tfsdk:"enabled"`
	Protocol            types.String   `tfsdk:"protocol"`
	SrcNetworkConfType  types.String   `tfsdk:"src_network_conf_type"`
	SrcAddress          types.String   `tfsdk:"src_address"`
	SrcFirewallGroupIDs types.List     `tfsdk:"src_firewall_group_ids"`
	DstNetworkConfType  types.String   `tfsdk:"dst_network_conf_type"`
	DstAddress          types.String   `tfsdk:"dst_address"`
	DstFirewallGroupIDs types.List     `tfsdk:"dst_firewall_group_ids"`
	DstPort             types.String   `tfsdk:"dst_port"`
	Logging             types.Bool     `tfsdk:"logging"`
	StateNew            types.Bool     `tfsdk:"state_new"`
	StateEstablished    types.Bool     `tfsdk:"state_established"`
	StateRelated        types.Bool     `tfsdk:"state_related"`
	StateInvalid        types.Bool     `tfsdk:"state_invalid"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

func NewFirewallRuleResource() resource.Resource {
	return &FirewallRuleResource{}
}

func (r *FirewallRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *FirewallRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi legacy firewall rule.",
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
				Description: "The unique identifier of the firewall rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the firewall rule is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall rule.",
				Required:    true,
			},
			"ruleset": schema.StringAttribute{
				Description: "The ruleset this rule belongs to. Valid values: 'WAN_IN', 'WAN_OUT', 'WAN_LOCAL', " +
					"'LAN_IN', 'LAN_OUT', 'LAN_LOCAL', 'GUEST_IN', 'GUEST_OUT', 'GUEST_LOCAL', " +
					"'WANv6_IN', 'WANv6_OUT', 'WANv6_LOCAL', 'LANv6_IN', 'LANv6_OUT', 'LANv6_LOCAL', " +
					"'GUESTv6_IN', 'GUESTv6_OUT', 'GUESTv6_LOCAL'.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"WAN_IN", "WAN_OUT", "WAN_LOCAL",
						"LAN_IN", "LAN_OUT", "LAN_LOCAL",
						"GUEST_IN", "GUEST_OUT", "GUEST_LOCAL",
						"WANv6_IN", "WANv6_OUT", "WANv6_LOCAL",
						"LANv6_IN", "LANv6_OUT", "LANv6_LOCAL",
						"GUESTv6_IN", "GUESTv6_OUT", "GUESTv6_LOCAL",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				Description: "The action to take. Valid values: 'accept', 'drop', 'reject'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("accept", "drop", "reject"),
				},
			},
			"rule_index": schema.Int64Attribute{
				Description: "The index/priority of the rule (lower numbers are evaluated first).",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol to match. Valid values: 'all', 'tcp', 'udp', 'tcp_udp', 'icmp'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "tcp", "udp", "tcp_udp", "icmp"),
				},
			},
			"src_network_conf_type": schema.StringAttribute{
				Description: "Source network configuration type. Valid values: 'ADDRv4', 'NETv4'.",
				Optional:    true,
			},
			"src_address": schema.StringAttribute{
				Description: "Source IP address or CIDR.",
				Optional:    true,
			},
			"src_firewall_group_ids": schema.ListAttribute{
				Description: "List of source firewall group IDs.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"dst_network_conf_type": schema.StringAttribute{
				Description: "Destination network configuration type. Valid values: 'ADDRv4', 'NETv4'.",
				Optional:    true,
			},
			"dst_address": schema.StringAttribute{
				Description: "Destination IP address or CIDR.",
				Optional:    true,
			},
			"dst_firewall_group_ids": schema.ListAttribute{
				Description: "List of destination firewall group IDs.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"dst_port": schema.StringAttribute{
				Description: "Destination port or port range (e.g., '80' or '8080-8090').",
				Optional:    true,
			},
			"logging": schema.BoolAttribute{
				Description: "Whether to log matching packets. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"state_new": schema.BoolAttribute{
				Description: "Match new connections. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"state_established": schema.BoolAttribute{
				Description: "Match established connections. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"state_related": schema.BoolAttribute{
				Description: "Match related connections. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"state_invalid": schema.BoolAttribute{
				Description: "Match invalid connections. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *FirewallRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallRuleResourceModel

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

	// Convert plan to SDK struct
	rule := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the firewall rule
	created, err := r.client.CreateFirewallRule(ctx, rule)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "firewall rule")
		return
	}

	// Update state with response
	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallRuleResourceModel

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

	// Get the firewall rule
	rule, err := r.client.GetFirewallRule(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "firewall rule")
		return
	}

	// Update state with response
	resp.Diagnostics.Append(r.sdkToState(ctx, rule, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FirewallRuleResourceModel

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

	var state FirewallRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert plan to SDK struct
	rule := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID and SiteID from state
	rule.ID = state.ID.ValueString()
	rule.SiteID = state.SiteID.ValueString()

	// Update the firewall rule
	updated, err := r.client.UpdateFirewallRule(ctx, state.ID.ValueString(), rule)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "firewall rule")
		return
	}

	// Update state with response
	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallRuleResourceModel

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

	// Delete the firewall rule
	err := r.client.DeleteFirewallRule(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "firewall rule")
		return
	}
}

func (r *FirewallRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// planToSDK converts the Terraform plan to an SDK FirewallRule struct.
func (r *FirewallRuleResource) planToSDK(ctx context.Context, plan *FirewallRuleResourceModel, diags *diag.Diagnostics) *unifi.FirewallRule {
	rule := &unifi.FirewallRule{
		Name:             plan.Name.ValueString(),
		Ruleset:          plan.Ruleset.ValueString(),
		Action:           plan.Action.ValueString(),
		RuleIndex:        intPtr(plan.RuleIndex.ValueInt64()),
		Enabled:          boolPtr(plan.Enabled.ValueBool()),
		Protocol:         plan.Protocol.ValueString(),
		Logging:          boolPtr(plan.Logging.ValueBool()),
		StateNew:         boolPtr(plan.StateNew.ValueBool()),
		StateEstablished: boolPtr(plan.StateEstablished.ValueBool()),
		StateRelated:     boolPtr(plan.StateRelated.ValueBool()),
		StateInvalid:     boolPtr(plan.StateInvalid.ValueBool()),
	}

	if !plan.SrcNetworkConfType.IsNull() {
		rule.SrcNetworkConfType = plan.SrcNetworkConfType.ValueString()
	}

	if !plan.SrcAddress.IsNull() {
		rule.SrcAddress = plan.SrcAddress.ValueString()
	}

	if !plan.SrcFirewallGroupIDs.IsNull() {
		var ids []string
		diags.Append(plan.SrcFirewallGroupIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil
		}
		rule.SrcFirewallGroupIDs = ids
	}

	if !plan.DstNetworkConfType.IsNull() {
		rule.DstNetworkConfType = plan.DstNetworkConfType.ValueString()
	}

	if !plan.DstAddress.IsNull() {
		rule.DstAddress = plan.DstAddress.ValueString()
	}

	if !plan.DstFirewallGroupIDs.IsNull() {
		var ids []string
		diags.Append(plan.DstFirewallGroupIDs.ElementsAs(ctx, &ids, false)...)
		if diags.HasError() {
			return nil
		}
		rule.DstFirewallGroupIDs = ids
	}

	if !plan.DstPort.IsNull() {
		rule.DstPort = plan.DstPort.ValueString()
	}

	return rule
}

// sdkToState updates the Terraform state from an SDK FirewallRule struct.
func (r *FirewallRuleResource) sdkToState(ctx context.Context, rule *unifi.FirewallRule, state *FirewallRuleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(rule.ID)
	state.SiteID = types.StringValue(rule.SiteID)
	state.Name = types.StringValue(rule.Name)
	state.Ruleset = types.StringValue(rule.Ruleset)
	state.Action = types.StringValue(rule.Action)

	if rule.RuleIndex != nil {
		state.RuleIndex = types.Int64Value(int64(*rule.RuleIndex))
	}

	state.Enabled = types.BoolValue(derefBool(rule.Enabled))
	state.Protocol = types.StringValue(rule.Protocol)

	if rule.SrcNetworkConfType != "" {
		state.SrcNetworkConfType = types.StringValue(rule.SrcNetworkConfType)
	} else {
		state.SrcNetworkConfType = types.StringNull()
	}

	if rule.SrcAddress != "" {
		state.SrcAddress = types.StringValue(rule.SrcAddress)
	} else {
		state.SrcAddress = types.StringNull()
	}

	if len(rule.SrcFirewallGroupIDs) > 0 {
		srcList, d := types.ListValueFrom(ctx, types.StringType, rule.SrcFirewallGroupIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.SrcFirewallGroupIDs = srcList
	} else {
		state.SrcFirewallGroupIDs = types.ListNull(types.StringType)
	}

	if rule.DstNetworkConfType != "" {
		state.DstNetworkConfType = types.StringValue(rule.DstNetworkConfType)
	} else {
		state.DstNetworkConfType = types.StringNull()
	}

	if rule.DstAddress != "" {
		state.DstAddress = types.StringValue(rule.DstAddress)
	} else {
		state.DstAddress = types.StringNull()
	}

	if len(rule.DstFirewallGroupIDs) > 0 {
		dstList, d := types.ListValueFrom(ctx, types.StringType, rule.DstFirewallGroupIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.DstFirewallGroupIDs = dstList
	} else {
		state.DstFirewallGroupIDs = types.ListNull(types.StringType)
	}

	if rule.DstPort != "" {
		state.DstPort = types.StringValue(rule.DstPort)
	} else {
		state.DstPort = types.StringNull()
	}

	state.Logging = types.BoolValue(derefBool(rule.Logging))
	state.StateNew = types.BoolValue(derefBool(rule.StateNew))
	state.StateEstablished = types.BoolValue(derefBool(rule.StateEstablished))
	state.StateRelated = types.BoolValue(derefBool(rule.StateRelated))
	state.StateInvalid = types.BoolValue(derefBool(rule.StateInvalid))

	return diags
}
