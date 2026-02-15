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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct {
	client *AutoLoginClient
}

type UserResourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	SiteID                types.String   `tfsdk:"site_id"`
	MAC                   types.String   `tfsdk:"mac"`
	Name                  types.String   `tfsdk:"name"`
	Note                  types.String   `tfsdk:"note"`
	Noted                 types.Bool     `tfsdk:"noted"`
	UseFixedIP            types.Bool     `tfsdk:"use_fixed_ip"`
	FixedIP               types.String   `tfsdk:"fixed_ip"`
	NetworkID             types.String   `tfsdk:"network_id"`
	LocalDnsRecord        types.String   `tfsdk:"local_dns_record"`
	LocalDnsRecordEnabled types.Bool     `tfsdk:"local_dns_record_enabled"`
	UsergroupID           types.String   `tfsdk:"usergroup_id"`
	Blocked               types.Bool     `tfsdk:"blocked"`
	IP                    types.String   `tfsdk:"ip"`
	Hostname              types.String   `tfsdk:"hostname"`
	OUI                   types.String   `tfsdk:"oui"`
	FirstSeen             types.Int64    `tfsdk:"first_seen"`
	LastSeen              types.Int64    `tfsdk:"last_seen"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi user (client device record). Users represent network clients with " +
			"DHCP reservations, fixed IPs, device names, local DNS records, blocking, and user group assignments.",
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
				Description: "The unique identifier of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_id": schema.StringAttribute{
				Description: "The site ID where the user is created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mac": schema.StringAttribute{
				Description: "The MAC address of the client device (format: aa:bb:cc:dd:ee:ff). " +
					"Changing this forces recreation of the resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A friendly name for the client device.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"note": schema.StringAttribute{
				Description: "Notes or description for the client device.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"noted": schema.BoolAttribute{
				Description: "Whether the device has a note. Automatically set to true when note is provided.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"use_fixed_ip": schema.BoolAttribute{
				Description: "Enable DHCP reservation for this device. Must be true for fixed_ip to take effect.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"fixed_ip": schema.StringAttribute{
				Description: "The fixed IP address for DHCP reservation. Requires use_fixed_ip to be true.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				Description: "The network ID for the fixed IP DHCP reservation.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_dns_record": schema.StringAttribute{
				Description: "A local DNS hostname record for this device.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_dns_record_enabled": schema.BoolAttribute{
				Description: "Whether the local DNS record is enabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"usergroup_id": schema.StringAttribute{
				Description: "The user group ID for bandwidth limiting (QoS profile assignment).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"blocked": schema.BoolAttribute{
				Description: "Whether the device is blocked from network access.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Description: "The current IP address of the device (read-only, assigned by the controller).",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the device from DHCP (read-only).",
				Computed:    true,
			},
			"oui": schema.StringAttribute{
				Description: "The manufacturer OUI code derived from the MAC address (read-only).",
				Computed:    true,
			},
			"first_seen": schema.Int64Attribute{
				Description: "Unix timestamp when the device was first seen (read-only).",
				Computed:    true,
			},
			"last_seen": schema.Int64Attribute{
				Description: "Unix timestamp when the device was last seen (read-only).",
				Computed:    true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel

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

	user := r.planToSDK(&plan)

	created, err := r.client.CreateUser(ctx, user)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "user")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel

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

	user, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "user")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(user, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel

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

	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := r.planToSDK(&plan)
	user.ID = state.ID.ValueString()
	user.SiteID = state.SiteID.ValueString()

	updated, err := r.client.UpdateUser(ctx, state.ID.ValueString(), user)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "user")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel

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

	err := r.client.DeleteUser(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "user")
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UserResource) planToSDK(plan *UserResourceModel) *unifi.User {
	user := &unifi.User{
		MAC: plan.MAC.ValueString(),
	}

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		user.Name = plan.Name.ValueString()
	}

	if !plan.Note.IsNull() && !plan.Note.IsUnknown() {
		user.Note = plan.Note.ValueString()
	}

	if !plan.Noted.IsNull() && !plan.Noted.IsUnknown() {
		user.Noted = boolPtr(plan.Noted.ValueBool())
	}

	if !plan.UseFixedIP.IsNull() && !plan.UseFixedIP.IsUnknown() {
		user.UseFixedIP = boolPtr(plan.UseFixedIP.ValueBool())
	}

	if !plan.FixedIP.IsNull() && !plan.FixedIP.IsUnknown() {
		user.FixedIP = plan.FixedIP.ValueString()
	}

	if !plan.NetworkID.IsNull() && !plan.NetworkID.IsUnknown() {
		user.NetworkID = plan.NetworkID.ValueString()
	}

	if !plan.LocalDnsRecord.IsNull() && !plan.LocalDnsRecord.IsUnknown() {
		user.LocalDnsRecord = plan.LocalDnsRecord.ValueString()
	}

	if !plan.LocalDnsRecordEnabled.IsNull() && !plan.LocalDnsRecordEnabled.IsUnknown() {
		user.LocalDnsRecordEnabled = boolPtr(plan.LocalDnsRecordEnabled.ValueBool())
	}

	if !plan.UsergroupID.IsNull() && !plan.UsergroupID.IsUnknown() {
		user.UsergroupID = plan.UsergroupID.ValueString()
	}

	if !plan.Blocked.IsNull() && !plan.Blocked.IsUnknown() {
		user.Blocked = boolPtr(plan.Blocked.ValueBool())
	}

	return user
}

func (r *UserResource) sdkToState(user *unifi.User, state *UserResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(user.ID)
	state.SiteID = types.StringValue(user.SiteID)
	state.MAC = types.StringValue(user.MAC)
	state.Name = stringValueOrNull(user.Name)
	state.Note = stringValueOrNull(user.Note)

	if user.Noted != nil {
		state.Noted = types.BoolValue(*user.Noted)
	} else {
		state.Noted = types.BoolNull()
	}

	if user.UseFixedIP != nil {
		state.UseFixedIP = types.BoolValue(*user.UseFixedIP)
	} else {
		state.UseFixedIP = types.BoolNull()
	}

	state.FixedIP = stringValueOrNull(user.FixedIP)
	state.NetworkID = stringValueOrNull(user.NetworkID)
	state.LocalDnsRecord = stringValueOrNull(user.LocalDnsRecord)

	if user.LocalDnsRecordEnabled != nil {
		state.LocalDnsRecordEnabled = types.BoolValue(*user.LocalDnsRecordEnabled)
	} else {
		state.LocalDnsRecordEnabled = types.BoolNull()
	}

	state.UsergroupID = stringValueOrNull(user.UsergroupID)

	if user.Blocked != nil {
		state.Blocked = types.BoolValue(*user.Blocked)
	} else {
		state.Blocked = types.BoolNull()
	}

	state.IP = stringValueOrNull(user.IP)
	state.Hostname = stringValueOrNull(user.Hostname)
	state.OUI = stringValueOrNull(user.OUI)

	if user.FirstSeen != nil {
		state.FirstSeen = types.Int64Value(*user.FirstSeen)
	} else {
		state.FirstSeen = types.Int64Null()
	}

	if user.LastSeen != nil {
		state.LastSeen = types.Int64Value(*user.LastSeen)
	} else {
		state.LastSeen = types.Int64Null()
	}

	return diags
}
