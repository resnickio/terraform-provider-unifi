package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &FirewallZoneResource{}
	_ resource.ResourceWithImportState = &FirewallZoneResource{}
)

type FirewallZoneResource struct {
	client *AutoLoginClient
}

type FirewallZoneResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ZoneKey    types.String `tfsdk:"zone_key"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
}

func NewFirewallZoneResource() resource.Resource {
	return &FirewallZoneResource{}
}

func (r *FirewallZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_zone"
}

func (r *FirewallZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi firewall zone (v2 zone-based firewall).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the firewall zone.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the firewall zone.",
				Required:    true,
			},
			"zone_key": schema.StringAttribute{
				Description: "The zone key for built-in zones. Valid values: 'internal', 'external', 'gateway', 'vpn', 'hotspot', 'dmz'. Leave empty for custom zones.",
				Optional:    true,
			},
			"network_ids": schema.ListAttribute{
				Description: "List of network IDs assigned to this zone.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *FirewallZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallZoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateFirewallZone(ctx, zone)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "firewall zone")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.client.GetFirewallZone(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "firewall zone")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, zone, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FirewallZoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FirewallZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := r.planToSDK(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	zone.ID = state.ID.ValueString()

	updated, err := r.client.UpdateFirewallZone(ctx, state.ID.ValueString(), zone)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "firewall zone")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFirewallZone(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "firewall zone")
		return
	}
}

func (r *FirewallZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *FirewallZoneResource) planToSDK(ctx context.Context, plan *FirewallZoneResourceModel, diags *diag.Diagnostics) *unifi.FirewallZone {
	zone := &unifi.FirewallZone{
		Name: plan.Name.ValueString(),
	}

	if !plan.ZoneKey.IsNull() && !plan.ZoneKey.IsUnknown() {
		zoneKey := plan.ZoneKey.ValueString()
		zone.ZoneKey = &zoneKey
	}

	if !plan.NetworkIDs.IsNull() && !plan.NetworkIDs.IsUnknown() {
		var networkIDs []string
		diags.Append(plan.NetworkIDs.ElementsAs(ctx, &networkIDs, false)...)
		if diags.HasError() {
			return nil
		}
		zone.NetworkIDs = networkIDs
	}

	return zone
}

func (r *FirewallZoneResource) sdkToState(ctx context.Context, zone *unifi.FirewallZone, state *FirewallZoneResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(zone.ID)
	state.Name = types.StringValue(zone.Name)

	if zone.ZoneKey != nil && *zone.ZoneKey != "" {
		state.ZoneKey = types.StringValue(*zone.ZoneKey)
	} else {
		state.ZoneKey = types.StringNull()
	}

	if len(zone.NetworkIDs) > 0 {
		networkIDsList, d := types.ListValueFrom(ctx, types.StringType, zone.NetworkIDs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		state.NetworkIDs = networkIDsList
	} else {
		state.NetworkIDs = types.ListNull(types.StringType)
	}

	return diags
}
