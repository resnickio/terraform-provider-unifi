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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var (
	_ resource.Resource                = &StaticDNSResource{}
	_ resource.ResourceWithImportState = &StaticDNSResource{}
)

type StaticDNSResource struct {
	client *AutoLoginClient
}

type StaticDNSResourceModel struct {
	ID         types.String   `tfsdk:"id"`
	Key        types.String   `tfsdk:"key"`
	Value      types.String   `tfsdk:"value"`
	RecordType types.String   `tfsdk:"record_type"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	TTL        types.Int64    `tfsdk:"ttl"`
	Port       types.Int64    `tfsdk:"port"`
	Priority   types.Int64    `tfsdk:"priority"`
	Weight     types.Int64    `tfsdk:"weight"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

func NewStaticDNSResource() resource.Resource {
	return &StaticDNSResource{}
}

func (r *StaticDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_dns"
}

func (r *StaticDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a UniFi static DNS record.",
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
				Description: "The unique identifier of the static DNS record.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The hostname or domain name for the DNS record.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value for the DNS record (IP address, hostname, or other value depending on record type).",
				Required:    true,
			},
			"record_type": schema.StringAttribute{
				Description: "The DNS record type. Valid values: A, AAAA, CNAME, MX, NS, TXT, SRV.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CNAME", "MX", "NS", "TXT", "SRV"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the DNS record is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live in seconds for the DNS record.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port number for SRV records. Must be between 1 and 65535.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"priority": schema.Int64Attribute{
				Description: "Priority value for MX and SRV records.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"weight": schema.Int64Attribute{
				Description: "Weight value for SRV records.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *StaticDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StaticDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticDNSResourceModel

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

	created, err := r.client.CreateStaticDNS(ctx, dns)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "create", "static DNS record")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StaticDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticDNSResourceModel

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

	dns, err := r.client.GetStaticDNS(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		handleSDKError(&resp.Diagnostics, err, "read", "static DNS record")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, dns, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StaticDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticDNSResourceModel

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

	var state StaticDNSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dns := r.planToSDK(&plan)
	dns.ID = state.ID.ValueString()

	updated, err := r.client.UpdateStaticDNS(ctx, state.ID.ValueString(), dns)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "update", "static DNS record")
		return
	}

	resp.Diagnostics.Append(r.sdkToState(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StaticDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticDNSResourceModel

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

	err := r.client.DeleteStaticDNS(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		handleSDKError(&resp.Diagnostics, err, "delete", "static DNS record")
		return
	}
}

func (r *StaticDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *StaticDNSResource) planToSDK(plan *StaticDNSResourceModel) *unifi.StaticDNS {
	dns := &unifi.StaticDNS{
		Key:        plan.Key.ValueString(),
		Value:      plan.Value.ValueString(),
		RecordType: plan.RecordType.ValueString(),
		Enabled:    boolPtr(plan.Enabled.ValueBool()),
	}

	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		dns.TTL = intPtr(plan.TTL.ValueInt64())
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		dns.Port = intPtr(plan.Port.ValueInt64())
	}

	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		dns.Priority = intPtr(plan.Priority.ValueInt64())
	}

	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		dns.Weight = intPtr(plan.Weight.ValueInt64())
	}

	return dns
}

func (r *StaticDNSResource) sdkToState(ctx context.Context, dns *unifi.StaticDNS, state *StaticDNSResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(dns.ID)
	state.Key = types.StringValue(dns.Key)
	state.Value = types.StringValue(dns.Value)
	state.RecordType = types.StringValue(dns.RecordType)
	state.Enabled = types.BoolValue(derefBool(dns.Enabled))

	if dns.TTL != nil {
		state.TTL = types.Int64Value(int64(*dns.TTL))
	} else {
		state.TTL = types.Int64Null()
	}

	if dns.Port != nil {
		state.Port = types.Int64Value(int64(*dns.Port))
	} else {
		state.Port = types.Int64Null()
	}

	if dns.Priority != nil {
		state.Priority = types.Int64Value(int64(*dns.Priority))
	} else {
		state.Priority = types.Int64Null()
	}

	if dns.Weight != nil {
		state.Weight = types.Int64Value(int64(*dns.Weight))
	} else {
		state.Weight = types.Int64Null()
	}

	return diags
}
