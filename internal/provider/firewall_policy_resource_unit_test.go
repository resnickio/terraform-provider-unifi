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

func TestMatchingTargetTypeFor(t *testing.T) {
	cases := []struct {
		matchingTarget string
		want           string
	}{
		{matchingTarget: "IP", want: "SPECIFIC"},
		{matchingTarget: "NETWORK", want: "SPECIFIC"},
		{matchingTarget: "DOMAIN", want: "SPECIFIC"},
		{matchingTarget: "REGION", want: "SPECIFIC"},
		{matchingTarget: "PORT_GROUP", want: "OBJECT"},
		{matchingTarget: "ADDRESS_GROUP", want: "OBJECT"},
		{matchingTarget: "ANY", want: ""},
		{matchingTarget: "", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.matchingTarget, func(t *testing.T) {
			if got := matchingTargetTypeFor(tc.matchingTarget); got != tc.want {
				t.Fatalf("matchingTargetTypeFor(%q) = %q, want %q", tc.matchingTarget, got, tc.want)
			}
		})
	}
}
