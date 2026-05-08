package provider

import "testing"

func TestDeriveMatchingTarget(t *testing.T) {
	cases := []struct {
		name         string
		hasIPs       bool
		hasNetworkID bool
		want         string
	}{
		{name: "neither", hasIPs: false, hasNetworkID: false, want: "ANY"},
		{name: "ips only", hasIPs: true, hasNetworkID: false, want: "IP"},
		{name: "network only", hasIPs: false, hasNetworkID: true, want: "NETWORK"},
		{name: "ips beats network", hasIPs: true, hasNetworkID: true, want: "IP"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveMatchingTarget(tc.hasIPs, tc.hasNetworkID); got != tc.want {
				t.Fatalf("deriveMatchingTarget(%v, %v) = %q, want %q", tc.hasIPs, tc.hasNetworkID, got, tc.want)
			}
		})
	}
}

func TestDeriveCreateAllowRespond(t *testing.T) {
	cases := []struct {
		action string
		want   bool
	}{
		{action: "ALLOW", want: true},
		{action: "BLOCK", want: false},
		{action: "REJECT", want: false},
		{action: "", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			if got := deriveCreateAllowRespond(tc.action); got != tc.want {
				t.Fatalf("deriveCreateAllowRespond(%q) = %v, want %v", tc.action, got, tc.want)
			}
		})
	}
}

func TestMatchingTargetTypeFor(t *testing.T) {
	cases := []struct {
		matchingTarget string
		want           string
	}{
		// Probe-confirmed inline literals: SPECIFIC.
		{matchingTarget: "IP", want: "SPECIFIC"},
		{matchingTarget: "NETWORK", want: "SPECIFIC"},
		{matchingTarget: "REGION", want: "SPECIFIC"},
		// Object reference preserved from prior probe.
		{matchingTarget: "IID", want: "OBJECT"},
		// ANY and unset: empty.
		{matchingTarget: "ANY", want: ""},
		{matchingTarget: "", want: ""},
		// Identity-aware values added in SDK v0.12.0. matching_target_type
		// requirements unknown — fall through to "" until probed.
		{matchingTarget: "CLIENT", want: ""},
		{matchingTarget: "EXTERNAL_SOURCE", want: ""},
		{matchingTarget: "MAC", want: ""},
		{matchingTarget: "USER_IDENTITY", want: ""},
		{matchingTarget: "USER_IDENTITY_ONE_CLICK_VPN", want: ""},
		{matchingTarget: "USER_IDENTITY_ONE_CLICK_WIFI", want: ""},
		{matchingTarget: "VPN_USER", want: ""},
		// Removed in v0.12.0 (controller rejects). The helper returns "" for
		// non-real values; the schema validator rejects them at plan time.
		{matchingTarget: "WEB", want: ""},
		{matchingTarget: "APP", want: ""},
		{matchingTarget: "APP_CATEGORY", want: ""},
		{matchingTarget: "DOMAIN", want: ""},
		{matchingTarget: "PORT_GROUP", want: ""},
		{matchingTarget: "ADDRESS_GROUP", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.matchingTarget, func(t *testing.T) {
			if got := matchingTargetTypeFor(tc.matchingTarget); got != tc.want {
				t.Fatalf("matchingTargetTypeFor(%q) = %q, want %q", tc.matchingTarget, got, tc.want)
			}
		})
	}
}
