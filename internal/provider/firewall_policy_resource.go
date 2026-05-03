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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                   = &FirewallPolicyResource{}
	_ resource.ResourceWithImportState    = &FirewallPolicyResource{}
	_ resource.ResourceWithModifyPlan     = &FirewallPolicyResource{}
	_ resource.ResourceWithValidateConfig = &FirewallPolicyResource{}
)

type FirewallPolicyResource struct {
	client *AutoLoginClient
}

type FirewallPolicyResourceModel struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Enabled             types.Bool     `tfsdk:"enabled"`
	Action              types.String   `tfsdk:"action"`
	Protocol            types.String   `tfsdk:"protocol"`
	IPVersion           types.String   `tfsdk:"ip_version"`
	Index               types.Int64    `tfsdk:"index"`
	Logging             types.Bool     `tfsdk:"logging"`
	ConnectionStateType types.String   `tfsdk:"connection_state_type"`
	ConnectionStates    types.Set      `tfsdk:"connection_states"`
	MatchIPSec          types.Bool     `tfsdk:"match_ipsec"`
	ICMPTypename        types.String   `tfsdk:"icmp_typename"`
	ICMPV6Typename      types.String   `tfsdk:"icmpv6_typename"`
	Source              types.Object   `tfsdk:"source"`
	Destination         types.Object   `tfsdk:"destination"`
	Schedule            types.Object   `tfsdk:"schedule"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

var endpointAttrTypes = map[string]attr.Type{
	"zone_id":         types.StringType,
	"matching_target": types.StringType,
	"ips":             types.SetType{ElemType: types.StringType},
	"mac":             types.StringType,
	"port":            types.StringType,
	"network_id":      types.StringType,
	"client_macs":     types.SetType{ElemType: types.StringType},
}

var scheduleAttrTypes = map[string]attr.Type{
	"mode":             types.StringType,
	"time_range_start": types.StringType,
	"time_range_end":   types.StringType,
	"days_of_week":     types.SetType{ElemType: types.StringType},
}

func NewFirewallPolicyResource() resource.Resource {
	return &FirewallPolicyResource{}
}

func (r *FirewallPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_policy"
}

func (r *FirewallPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi firewall policy (v2 zone-based firewall).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall policy.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the policy is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"action": schema.StringAttribute{
				Description: "The action to take. Valid values: 'ALLOW', 'BLOCK', 'REJECT'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW", "BLOCK", "REJECT"),
				},
			},
			"protocol": schema.StringAttribute{
				Description: "The protocol to match. Valid values: 'all', 'tcp_udp', 'tcp', 'udp', 'icmp', 'icmpv6'. Defaults to 'all'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "tcp_udp", "tcp", "udp", "icmp", "icmpv6"),
				},
			},
			"ip_version": schema.StringAttribute{
				Description: "The IP version to match. Valid values: 'BOTH', 'IPV4', 'IPV6'. Defaults to 'BOTH'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("BOTH"),
				Validators: []validator.String{
					stringvalidator.OneOf("BOTH", "IPV4", "IPV6"),
				},
			},
			"index": schema.Int64Attribute{
				Description: "The index/priority of the policy, assigned by the controller. Read-only — the controller determines ordering.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"logging": schema.BoolAttribute{
				Description: "Whether to log matching packets. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"connection_state_type": schema.StringAttribute{
				Description: "Connection state matching type. Valid values: 'ALL', 'RESPOND_ONLY', 'CUSTOM'. Defaults to 'ALL'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ALL"),
				Validators: []validator.String{
					stringvalidator.OneOf("ALL", "RESPOND_ONLY", "CUSTOM"),
				},
			},
			"connection_states": schema.SetAttribute{
				Description: "Set of connection states to match (when connection_state_type is 'CUSTOM').",
				Optional:    true,
				ElementType: types.StringType,
			},
			"match_ipsec": schema.BoolAttribute{
				Description: "Whether to match IPSec traffic. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"icmp_typename": schema.StringAttribute{
				Description: "ICMP type name (for ICMP protocol). Defaults to 'ANY'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ANY"),
			},
			"icmpv6_typename": schema.StringAttribute{
				Description: "ICMPv6 type name (for ICMPv6 protocol). Defaults to 'ANY'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ANY"),
			},
			"source": schema.SingleNestedAttribute{
				Description: "Source matching criteria.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"zone_id": schema.StringAttribute{
						Description: "The source zone ID.",
						Optional:    true,
					},
					"matching_target": schema.StringAttribute{
						Description: "Matching target type. Valid values: 'ANY', 'IP', 'NETWORK', 'DOMAIN', 'REGION', 'PORT_GROUP', 'ADDRESS_GROUP'. Auto-derived from sibling fields when unset: 'IP' if ips is non-empty, 'NETWORK' if network_id is non-empty, otherwise 'ANY'. (Unknown values from interpolation are treated as non-empty for derivation purposes.)",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANY", "IP", "NETWORK", "DOMAIN", "REGION", "PORT_GROUP", "ADDRESS_GROUP"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"ips": schema.SetAttribute{
						Description: "Set of IP addresses or CIDR ranges to match.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"mac": schema.StringAttribute{
						Description: "MAC address to match.",
						Optional:    true,
					},
					"port": schema.StringAttribute{
						Description: "Port or port range to match.",
						Optional:    true,
					},
					"network_id": schema.StringAttribute{
						Description: "Network ID to match.",
						Optional:    true,
					},
					"client_macs": schema.SetAttribute{
						Description: "Set of client MAC addresses to match.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"destination": schema.SingleNestedAttribute{
				Description: "Destination matching criteria.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"zone_id": schema.StringAttribute{
						Description: "The destination zone ID.",
						Optional:    true,
					},
					"matching_target": schema.StringAttribute{
						Description: "Matching target type. Valid values: 'ANY', 'IP', 'NETWORK', 'DOMAIN', 'REGION', 'PORT_GROUP', 'ADDRESS_GROUP'. Auto-derived from sibling fields when unset: 'IP' if ips is non-empty, 'NETWORK' if network_id is non-empty, otherwise 'ANY'. (Unknown values from interpolation are treated as non-empty for derivation purposes.)",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANY", "IP", "NETWORK", "DOMAIN", "REGION", "PORT_GROUP", "ADDRESS_GROUP"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"ips": schema.SetAttribute{
						Description: "Set of IP addresses or CIDR ranges to match.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"mac": schema.StringAttribute{
						Description: "MAC address to match.",
						Optional:    true,
					},
					"port": schema.StringAttribute{
						Description: "Port or port range to match.",
						Optional:    true,
					},
					"network_id": schema.StringAttribute{
						Description: "Network ID to match.",
						Optional:    true,
					},
					"client_macs": schema.SetAttribute{
						Description: "Set of client MAC addresses to match.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				Description: "Schedule configuration for when the policy is active.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Schedule mode. Valid values: 'ALWAYS', 'CUSTOM'.",
						Optional:    true,
					},
					"time_range_start": schema.StringAttribute{
						Description: "Start time in HH:MM format.",
						Optional:    true,
					},
					"time_range_end": schema.StringAttribute{
						Description: "End time in HH:MM format.",
						Optional:    true,
					},
					"days_of_week": schema.SetAttribute{
						Description: "Days of the week. Valid values: 'MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY'.",
						Optional:    true,
						ElementType: types.StringType,
					},
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

func (r *FirewallPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallPolicyResourceModel

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

	policy := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateFirewallPolicy(ctx, policy)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "firewall policy")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallPolicyResourceModel

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

	policy, err := r.client.GetFirewallPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "firewall policy")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, policy, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FirewallPolicyResourceModel

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

	var state FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	policy.ID = state.ID.ValueString()

	updated, err := r.client.UpdateFirewallPolicy(ctx, state.ID.ValueString(), policy)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "firewall policy")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallPolicyResourceModel

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

	err := r.client.DeleteFirewallPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "firewall policy")
		return
	}
}

func (r *FirewallPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan auto-derives source.matching_target and destination.matching_target
// from sibling fields when the user did not explicitly set them in config.
//
// UniFi's API silently discards `ips` and `network_id` when matching_target=ANY,
// so a static "ANY" default would let users lose data on a plan that looks
// cosmetic. Resolving here (instead of via a static default) means the plan
// diff shows the real matching_target before approval.
func (r *FirewallPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	r.deriveEndpointMatchingTarget(ctx, "source", req, resp)
	r.deriveEndpointMatchingTarget(ctx, "destination", req, resp)
}

func (r *FirewallPolicyResource) deriveEndpointMatchingTarget(ctx context.Context, endpoint string, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	endpointPath := path.Root(endpoint)

	var planned types.Object
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, endpointPath, &planned)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if planned.IsNull() || planned.IsUnknown() {
		return
	}

	var configured types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, endpointPath, &configured)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the user configured matching_target at all — even to a value that's
	// still unknown (e.g. matching_target = some_resource.attr) — preserve
	// their intent and let Terraform resolve it later. Auto-derivation only
	// fills in null values, never unknowns.
	if !configured.IsNull() && !configured.IsUnknown() {
		if mt, ok := configured.Attributes()["matching_target"].(types.String); ok && !mt.IsNull() {
			return
		}
	}

	plannedAttrs := planned.Attributes()

	hasIPs := false
	if v, ok := plannedAttrs["ips"].(types.Set); ok && !v.IsNull() {
		if v.IsUnknown() {
			hasIPs = true
		} else {
			hasIPs = len(v.Elements()) > 0
		}
	}

	hasNetworkID := false
	if v, ok := plannedAttrs["network_id"].(types.String); ok && !v.IsNull() {
		if v.IsUnknown() {
			hasNetworkID = true
		} else {
			hasNetworkID = v.ValueString() != ""
		}
	}

	derived := deriveMatchingTarget(hasIPs, hasNetworkID)

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, endpointPath.AtName("matching_target"), derived)...)
}

// deriveMatchingTarget picks the matching_target value from the populated
// sibling fields. Extracted so the precedence rules can be unit-tested
// without a controller round-trip.
func deriveMatchingTarget(hasIPs, hasNetworkID bool) string {
	switch {
	case hasIPs:
		return "IP"
	case hasNetworkID:
		return "NETWORK"
	default:
		return "ANY"
	}
}

// matchingTargetTypeFor maps a matching_target value to the matching_target_type
// the UniFi controller requires alongside it. The controller rejects an "IP"
// (or NETWORK / DOMAIN / REGION) match without an explicit "SPECIFIC" type, and
// rejects a *_GROUP match without an explicit "OBJECT" type. The provider
// derives this transparently — the field isn't exposed in the resource schema.
func matchingTargetTypeFor(matchingTarget string) string {
	switch matchingTarget {
	case "IP", "NETWORK", "DOMAIN", "REGION":
		return "SPECIFIC"
	case "PORT_GROUP", "ADDRESS_GROUP":
		return "OBJECT"
	default:
		return ""
	}
}

// ValidateConfig errors at plan time when the user explicitly sets
// matching_target="ANY" alongside ips or network_id. UniFi silently discards
// those fields under matching_target=ANY, so this combination is always wrong.
func (r *FirewallPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config FirewallPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.validateEndpointConfig("source", config.Source, &resp.Diagnostics)
	r.validateEndpointConfig("destination", config.Destination, &resp.Diagnostics)
}

func (r *FirewallPolicyResource) validateEndpointConfig(endpoint string, obj types.Object, diags *diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return
	}
	attrs := obj.Attributes()

	mt, ok := attrs["matching_target"].(types.String)
	if !ok || mt.IsNull() || mt.IsUnknown() || mt.ValueString() != "ANY" {
		return
	}

	// Unknown values (e.g. ips = var.foo where the variable hasn't resolved)
	// still count as "present" — the foot-gun applies just as much when the
	// values arrive at apply time as when they're known at plan time.
	hasIPs := false
	if v, ok := attrs["ips"].(types.Set); ok && !v.IsNull() {
		hasIPs = v.IsUnknown() || len(v.Elements()) > 0
	}
	hasNetworkID := false
	if v, ok := attrs["network_id"].(types.String); ok && !v.IsNull() {
		hasNetworkID = v.IsUnknown() || v.ValueString() != ""
	}

	if hasIPs {
		diags.AddAttributeError(
			path.Root(endpoint).AtName("matching_target"),
			"matching_target=\"ANY\" conflicts with ips",
			"The UniFi controller silently discards the ips field when matching_target=\"ANY\", which widens the policy. Either omit matching_target (it will be auto-derived to \"IP\") or set matching_target=\"IP\" explicitly.",
		)
	}
	if hasNetworkID {
		diags.AddAttributeError(
			path.Root(endpoint).AtName("matching_target"),
			"matching_target=\"ANY\" conflicts with network_id",
			"The UniFi controller silently discards the network_id field when matching_target=\"ANY\". Either omit matching_target (it will be auto-derived to \"NETWORK\") or set matching_target=\"NETWORK\" explicitly.",
		)
	}
}

func (r *FirewallPolicyResource) planToSDK(ctx context.Context, plan *FirewallPolicyResourceModel, diags *diag.Diagnostics) *unifi.FirewallPolicy {
	policy := &unifi.FirewallPolicy{
		Name:                plan.Name.ValueString(),
		Enabled:             boolPtr(plan.Enabled.ValueBool()),
		Action:              plan.Action.ValueString(),
		Protocol:            plan.Protocol.ValueString(),
		IPVersion:           plan.IPVersion.ValueString(),
		Logging:             boolPtr(plan.Logging.ValueBool()),
		ConnectionStateType: plan.ConnectionStateType.ValueString(),
		MatchIPSec:          boolPtr(plan.MatchIPSec.ValueBool()),
	}

	// Only send index on updates (when state has a value).
	// On create, let the controller assign it.
	if !plan.Index.IsNull() && !plan.Index.IsUnknown() {
		policy.Index = intPtr(plan.Index.ValueInt64())
	}

	if !plan.ConnectionStates.IsNull() && !plan.ConnectionStates.IsUnknown() {
		var states []string
		diags.Append(plan.ConnectionStates.ElementsAs(ctx, &states, false)...)
		if diags.HasError() {
			return nil
		}
		policy.ConnectionStates = states
	}

	if !plan.ICMPTypename.IsNull() && !plan.ICMPTypename.IsUnknown() {
		policy.ICMPTypename = plan.ICMPTypename.ValueString()
	}

	if !plan.ICMPV6Typename.IsNull() && !plan.ICMPV6Typename.IsUnknown() {
		policy.ICMPV6Typename = plan.ICMPV6Typename.ValueString()
	}

	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
		policy.Source = r.endpointFromObject(ctx, plan.Source, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		// Source is required by API - provide empty default
		policy.Source = &unifi.PolicyEndpoint{MatchingTarget: "ANY"}
	}

	if !plan.Destination.IsNull() && !plan.Destination.IsUnknown() {
		policy.Destination = r.endpointFromObject(ctx, plan.Destination, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		// Destination is required by API - provide empty default
		policy.Destination = &unifi.PolicyEndpoint{MatchingTarget: "ANY"}
	}

	if !plan.Schedule.IsNull() && !plan.Schedule.IsUnknown() {
		policy.Schedule = r.scheduleFromObject(ctx, plan.Schedule, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		// Schedule is required by API - provide default "always" schedule
		policy.Schedule = &unifi.PolicySchedule{Mode: "ALWAYS"}
	}

	return policy
}

func (r *FirewallPolicyResource) endpointFromObject(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *unifi.PolicyEndpoint {
	attrs := obj.Attributes()
	endpoint := &unifi.PolicyEndpoint{}

	if v, ok := attrs["zone_id"].(types.String); ok && !v.IsNull() {
		endpoint.ZoneID = v.ValueString()
	}
	if v, ok := attrs["matching_target"].(types.String); ok && !v.IsNull() {
		endpoint.MatchingTarget = v.ValueString()
		endpoint.MatchingTargetType = matchingTargetTypeFor(endpoint.MatchingTarget)
	}
	if v, ok := attrs["mac"].(types.String); ok && !v.IsNull() {
		endpoint.MAC = v.ValueString()
	}
	if v, ok := attrs["port"].(types.String); ok && !v.IsNull() {
		endpoint.Port = v.ValueString()
	}
	if v, ok := attrs["network_id"].(types.String); ok && !v.IsNull() {
		endpoint.NetworkID = v.ValueString()
	}
	if v, ok := attrs["ips"].(types.Set); ok && !v.IsNull() {
		var ips []string
		diags.Append(v.ElementsAs(ctx, &ips, false)...)
		if diags.HasError() {
			return nil
		}
		endpoint.IPs = ips
	}
	if v, ok := attrs["client_macs"].(types.Set); ok && !v.IsNull() {
		var macs []string
		diags.Append(v.ElementsAs(ctx, &macs, false)...)
		if diags.HasError() {
			return nil
		}
		endpoint.ClientMACs = macs
	}

	return endpoint
}

func (r *FirewallPolicyResource) scheduleFromObject(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *unifi.PolicySchedule {
	attrs := obj.Attributes()
	schedule := &unifi.PolicySchedule{}

	if v, ok := attrs["mode"].(types.String); ok && !v.IsNull() {
		schedule.Mode = v.ValueString()
	}
	if v, ok := attrs["time_range_start"].(types.String); ok && !v.IsNull() {
		schedule.TimeRangeStart = v.ValueString()
	}
	if v, ok := attrs["time_range_end"].(types.String); ok && !v.IsNull() {
		schedule.TimeRangeEnd = v.ValueString()
	}
	if v, ok := attrs["days_of_week"].(types.Set); ok && !v.IsNull() {
		var days []string
		diags.Append(v.ElementsAs(ctx, &days, false)...)
		if diags.HasError() {
			return nil
		}
		schedule.DaysOfWeek = days
	}

	return schedule
}

func (r *FirewallPolicyResource) sdkToState(ctx context.Context, policy *unifi.FirewallPolicy, state *FirewallPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	state.Enabled = types.BoolValue(derefBool(policy.Enabled))
	state.Action = types.StringValue(policy.Action)
	state.Protocol = types.StringValue(policy.Protocol)
	state.IPVersion = types.StringValue(policy.IPVersion)

	if policy.Index != nil {
		state.Index = types.Int64Value(int64(*policy.Index))
	} else {
		state.Index = types.Int64Null()
	}

	state.Logging = types.BoolValue(derefBool(policy.Logging))
	state.ConnectionStateType = types.StringValue(policy.ConnectionStateType)

	if len(policy.ConnectionStates) > 0 {
		statesSet, d := types.SetValueFrom(ctx, types.StringType, policy.ConnectionStates)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.ConnectionStates = statesSet
	} else {
		state.ConnectionStates = types.SetNull(types.StringType)
	}

	state.MatchIPSec = types.BoolValue(derefBool(policy.MatchIPSec))

	if policy.ICMPTypename != "" {
		state.ICMPTypename = types.StringValue(policy.ICMPTypename)
	} else {
		state.ICMPTypename = types.StringValue("ANY")
	}

	if policy.ICMPV6Typename != "" {
		state.ICMPV6Typename = types.StringValue(policy.ICMPV6Typename)
	} else {
		state.ICMPV6Typename = types.StringValue("ANY")
	}

	if policy.Source != nil {
		sourceObj, d := r.endpointToObject(ctx, policy.Source)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Source = sourceObj
	} else {
		state.Source = types.ObjectNull(endpointAttrTypes)
	}

	if policy.Destination != nil {
		destObj, d := r.endpointToObject(ctx, policy.Destination)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Destination = destObj
	} else {
		state.Destination = types.ObjectNull(endpointAttrTypes)
	}

	if policy.Schedule != nil {
		scheduleObj, d := r.scheduleToObject(ctx, policy.Schedule)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.Schedule = scheduleObj
	} else {
		state.Schedule = types.ObjectNull(scheduleAttrTypes)
	}

	return diags
}

func (r *FirewallPolicyResource) endpointToObject(ctx context.Context, endpoint *unifi.PolicyEndpoint) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var ipsVal types.Set
	if len(endpoint.IPs) > 0 {
		ipsSet, d := types.SetValueFrom(ctx, types.StringType, endpoint.IPs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(endpointAttrTypes), diags
		}
		ipsVal = ipsSet
	} else {
		ipsVal = types.SetNull(types.StringType)
	}

	var clientMacsVal types.Set
	if len(endpoint.ClientMACs) > 0 {
		macsSet, d := types.SetValueFrom(ctx, types.StringType, endpoint.ClientMACs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(endpointAttrTypes), diags
		}
		clientMacsVal = macsSet
	} else {
		clientMacsVal = types.SetNull(types.StringType)
	}

	matchingTarget := endpoint.MatchingTarget
	if matchingTarget == "" {
		matchingTarget = "ANY"
	}

	attrs := map[string]attr.Value{
		"zone_id":         stringValueOrNull(endpoint.ZoneID),
		"matching_target": types.StringValue(matchingTarget),
		"ips":             ipsVal,
		"mac":             stringValueOrNull(endpoint.MAC),
		"port":            stringValueOrNull(endpoint.Port),
		"network_id":      stringValueOrNull(endpoint.NetworkID),
		"client_macs":     clientMacsVal,
	}

	obj, d := types.ObjectValue(endpointAttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}

func (r *FirewallPolicyResource) scheduleToObject(ctx context.Context, schedule *unifi.PolicySchedule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var daysVal types.Set
	if len(schedule.DaysOfWeek) > 0 {
		daysSet, d := types.SetValueFrom(ctx, types.StringType, schedule.DaysOfWeek)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(scheduleAttrTypes), diags
		}
		daysVal = daysSet
	} else {
		daysVal = types.SetNull(types.StringType)
	}

	attrs := map[string]attr.Value{
		"mode":             stringValueOrNull(schedule.Mode),
		"time_range_start": stringValueOrNull(schedule.TimeRangeStart),
		"time_range_end":   stringValueOrNull(schedule.TimeRangeEnd),
		"days_of_week":     daysVal,
	}

	obj, d := types.ObjectValue(scheduleAttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}
