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
		// Inline literal values: SPECIFIC.
		{matchingTarget: "IP", want: "SPECIFIC"},
		{matchingTarget: "NETWORK", want: "SPECIFIC"},
		{matchingTarget: "REGION", want: "SPECIFIC"},
		{matchingTarget: "WEB", want: "SPECIFIC"},
		// Object references: OBJECT.
		{matchingTarget: "APP", want: "OBJECT"},
		{matchingTarget: "APP_CATEGORY", want: "OBJECT"},
		{matchingTarget: "IID", want: "OBJECT"},
		// ANY and unset: empty (no companion type needed).
		{matchingTarget: "ANY", want: ""},
		{matchingTarget: "", want: ""},
		// Pre-v0.9.0 values that used to be in this map but the controller
		// never accepted. The helper returns "" for any non-real enum value;
		// the schema validator now rejects these at plan time.
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
