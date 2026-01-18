package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var trafficTargetAttrTypes = map[string]attr.Type{
	"client_mac": types.StringType,
	"type":       types.StringType,
	"network_id": types.StringType,
}

var trafficDomainAttrTypes = map[string]attr.Type{
	"domain":      types.StringType,
	"description": types.StringType,
	"ports":       types.SetType{ElemType: types.Int64Type},
}

var trafficScheduleAttrTypes = map[string]attr.Type{
	"mode":             types.StringType,
	"time_range_start": types.StringType,
	"time_range_end":   types.StringType,
	"days_of_week":     types.SetType{ElemType: types.StringType},
}

var trafficBandwidthAttrTypes = map[string]attr.Type{
	"download_limit_kbps": types.Int64Type,
	"upload_limit_kbps":   types.Int64Type,
	"enabled":             types.BoolType,
}

func trafficTargetsFromList(ctx context.Context, list types.List, diags *diag.Diagnostics) []unifi.TrafficRuleTarget {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var targets []unifi.TrafficRuleTarget
	elements := list.Elements()

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attrs := obj.Attributes()
		target := unifi.TrafficRuleTarget{}

		if v, ok := attrs["client_mac"].(types.String); ok && !v.IsNull() {
			target.ClientMAC = v.ValueString()
		}
		if v, ok := attrs["type"].(types.String); ok && !v.IsNull() {
			target.Type = v.ValueString()
		}
		if v, ok := attrs["network_id"].(types.String); ok && !v.IsNull() {
			target.NetworkID = v.ValueString()
		}

		targets = append(targets, target)
	}

	return targets
}

func trafficTargetsToList(ctx context.Context, targets []unifi.TrafficRuleTarget) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(targets) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: trafficTargetAttrTypes}), diags
	}

	var elements []attr.Value
	for _, target := range targets {
		attrs := map[string]attr.Value{
			"client_mac": stringValueOrNull(target.ClientMAC),
			"type":       stringValueOrNull(target.Type),
			"network_id": stringValueOrNull(target.NetworkID),
		}
		obj, d := types.ObjectValue(trafficTargetAttrTypes, attrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: trafficTargetAttrTypes}), diags
		}
		elements = append(elements, obj)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: trafficTargetAttrTypes}, elements)
	diags.Append(d...)
	return list, diags
}

func trafficDomainsFromList(ctx context.Context, list types.List, diags *diag.Diagnostics) []unifi.TrafficDomain {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var domains []unifi.TrafficDomain
	elements := list.Elements()

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}

		attrs := obj.Attributes()
		domain := unifi.TrafficDomain{}

		if v, ok := attrs["domain"].(types.String); ok && !v.IsNull() {
			domain.Domain = v.ValueString()
		}
		if v, ok := attrs["description"].(types.String); ok && !v.IsNull() {
			domain.Description = v.ValueString()
		}
		if v, ok := attrs["ports"].(types.Set); ok && !v.IsNull() {
			var ports []int64
			d := v.ElementsAs(ctx, &ports, false)
			diags.Append(d...)
			for _, p := range ports {
				domain.Ports = append(domain.Ports, int(p))
			}
		}

		domains = append(domains, domain)
	}

	return domains
}

func trafficDomainsToList(ctx context.Context, domains []unifi.TrafficDomain) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(domains) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: trafficDomainAttrTypes}), diags
	}

	var elements []attr.Value
	for _, domain := range domains {
		var portsSet types.Set
		if len(domain.Ports) > 0 {
			var ports []attr.Value
			for _, p := range domain.Ports {
				ports = append(ports, types.Int64Value(int64(p)))
			}
			var d diag.Diagnostics
			portsSet, d = types.SetValue(types.Int64Type, ports)
			diags.Append(d...)
		} else {
			portsSet = types.SetNull(types.Int64Type)
		}

		attrs := map[string]attr.Value{
			"domain":      types.StringValue(domain.Domain),
			"description": stringValueOrNull(domain.Description),
			"ports":       portsSet,
		}
		obj, d := types.ObjectValue(trafficDomainAttrTypes, attrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: trafficDomainAttrTypes}), diags
		}
		elements = append(elements, obj)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: trafficDomainAttrTypes}, elements)
	diags.Append(d...)
	return list, diags
}

func trafficScheduleFromObject(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *unifi.PolicySchedule {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

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
		d := v.ElementsAs(ctx, &days, false)
		diags.Append(d...)
		schedule.DaysOfWeek = days
	}

	return schedule
}

func trafficScheduleToObject(ctx context.Context, schedule *unifi.PolicySchedule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if schedule == nil {
		return types.ObjectNull(trafficScheduleAttrTypes), diags
	}

	var daysSet types.Set
	if len(schedule.DaysOfWeek) > 0 {
		var days []attr.Value
		for _, d := range schedule.DaysOfWeek {
			days = append(days, types.StringValue(d))
		}
		var d diag.Diagnostics
		daysSet, d = types.SetValue(types.StringType, days)
		diags.Append(d...)
	} else {
		daysSet = types.SetNull(types.StringType)
	}

	attrs := map[string]attr.Value{
		"mode":             stringValueOrNull(schedule.Mode),
		"time_range_start": stringValueOrNull(schedule.TimeRangeStart),
		"time_range_end":   stringValueOrNull(schedule.TimeRangeEnd),
		"days_of_week":     daysSet,
	}

	obj, d := types.ObjectValue(trafficScheduleAttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}

func trafficBandwidthFromObject(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *unifi.TrafficBandwidth {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	attrs := obj.Attributes()
	bw := &unifi.TrafficBandwidth{}

	if v, ok := attrs["download_limit_kbps"].(types.Int64); ok && !v.IsNull() {
		val := int(v.ValueInt64())
		bw.DownloadLimitKbps = &val
	}
	if v, ok := attrs["upload_limit_kbps"].(types.Int64); ok && !v.IsNull() {
		val := int(v.ValueInt64())
		bw.UploadLimitKbps = &val
	}
	if v, ok := attrs["enabled"].(types.Bool); ok && !v.IsNull() {
		val := v.ValueBool()
		bw.Enabled = &val
	}

	return bw
}

func trafficBandwidthToObject(ctx context.Context, bw *unifi.TrafficBandwidth) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if bw == nil {
		return types.ObjectNull(trafficBandwidthAttrTypes), diags
	}

	var downloadLimit types.Int64
	if bw.DownloadLimitKbps != nil {
		downloadLimit = types.Int64Value(int64(*bw.DownloadLimitKbps))
	} else {
		downloadLimit = types.Int64Null()
	}

	var uploadLimit types.Int64
	if bw.UploadLimitKbps != nil {
		uploadLimit = types.Int64Value(int64(*bw.UploadLimitKbps))
	} else {
		uploadLimit = types.Int64Null()
	}

	var enabled types.Bool
	if bw.Enabled != nil {
		enabled = types.BoolValue(*bw.Enabled)
	} else {
		enabled = types.BoolNull()
	}

	attrs := map[string]attr.Value{
		"download_limit_kbps": downloadLimit,
		"upload_limit_kbps":   uploadLimit,
		"enabled":             enabled,
	}

	obj, d := types.ObjectValue(trafficBandwidthAttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}

func isEmptySchedule(schedule *unifi.PolicySchedule) bool {
	if schedule == nil {
		return true
	}
	return schedule.Mode == "" && schedule.TimeRangeStart == "" &&
		schedule.TimeRangeEnd == "" && len(schedule.DaysOfWeek) == 0
}

func isEmptyBandwidthLimit(bw *unifi.TrafficBandwidth) bool {
	if bw == nil {
		return true
	}
	enabled := bw.Enabled == nil || !*bw.Enabled
	download := bw.DownloadLimitKbps == nil || *bw.DownloadLimitKbps == 0
	upload := bw.UploadLimitKbps == nil || *bw.UploadLimitKbps == 0
	return enabled && download && upload
}
