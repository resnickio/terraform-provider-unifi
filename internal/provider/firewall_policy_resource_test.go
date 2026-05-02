package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallPolicyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy", "ALLOW"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "ALLOW"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "index"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "all"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "BOTH"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_block(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-block", "BLOCK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-block"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "BLOCK"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_full("tf-acc-test-policy-full"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-full"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "ALLOW"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "IPV4"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "index"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "logging", "true"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-update", "ALLOW"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-update"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "ALLOW"),
				),
			},
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-updated", "BLOCK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-updated"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "BLOCK"),
				),
			},
		},
	})
}

func TestAccFirewallPolicyResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_disabled("tf-acc-test-policy-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-disabled"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_tcp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_protocol("tf-acc-test-policy-tcp", "tcp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-tcp"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "tcp"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_udp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_protocol("tf-acc-test-policy-udp", "udp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-udp"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "udp"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_ipv4Only(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_ipVersion("tf-acc-test-policy-ipv4", "IPV4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-ipv4"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "IPV4"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_ipv6Only(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_ipVersion("tf-acc-test-policy-ipv6", "IPV6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-ipv6"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "IPV6"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_withSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_withSource("tf-acc-test-policy-src", "10.0.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-src"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.ips.#", "1"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.ips.0", "10.0.0.0/24"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_withDestination(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_withDestination("tf-acc-test-policy-dst", "192.168.1.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-dst"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.ips.#", "1"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.ips.0", "192.168.1.0/24"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_withPort(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_withPort("tf-acc-test-policy-port", "443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-port"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.port", "443"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_defaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-defaults", "ALLOW"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "all"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "BOTH"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "logging", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "connection_state_type", "ALL"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "match_ipsec", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "icmp_typename", "ANY"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "icmpv6_typename", "ANY"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.matching_target", "ANY"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.matching_target", "ANY"),
				),
			},
		},
	})
}

func TestAccFirewallPolicyResource_withZones(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_withZones("tf-acc-test-policy-zones"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-zones"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "source.zone_id"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "destination.zone_id"),
				),
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFirewallPolicyResourceConfig_basic(name, action string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = %q

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, action)
}

func testAccFirewallPolicyResourceConfig_full(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name       = %q
  action     = "ALLOW"
  protocol   = "tcp"
  ip_version = "IPV4"
  logging    = true
  enabled    = true

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name)
}

func testAccFirewallPolicyResourceConfig_disabled(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name    = %q
  action  = "ALLOW"
  enabled = false

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name)
}

func testAccFirewallPolicyResourceConfig_protocol(name, protocol string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name     = %q
  action   = "ALLOW"
  protocol = %q

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, protocol)
}

func testAccFirewallPolicyResourceConfig_ipVersion(name, ipVersion string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name       = %q
  action     = "ALLOW"
  ip_version = %q

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, ipVersion)
}

func testAccFirewallPolicyResourceConfig_withSource(name string, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "ALLOW"

  source = {
    zone_id         = unifi_firewall_zone.src.id
    matching_target = "IP"
    ips             = [%q]
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, sourceIP)
}

func testAccFirewallPolicyResourceConfig_withDestination(name string, destIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "ALLOW"

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id         = unifi_firewall_zone.dst.id
    matching_target = "IP"
    ips             = [%q]
  }
}
`, testAccProviderConfig, name, name, name, destIP)
}

func testAccFirewallPolicyResourceConfig_withPort(name string, port string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name     = %q
  action   = "ALLOW"
  protocol = "tcp"

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
    port    = %q
  }
}
`, testAccProviderConfig, name, name, name, port)
}

func testAccFirewallPolicyResourceConfig_withZones(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "source" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "destination" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "ALLOW"

  source = {
    zone_id = unifi_firewall_zone.source.id
  }

  destination = {
    zone_id = unifi_firewall_zone.destination.id
  }
}
`, testAccProviderConfig, name, name, name)
}

// Regression test for the silent data-loss bug where omitting matching_target
// caused the provider to send matching_target="ANY" alongside ips=[...], which
// the UniFi API silently strips. With ModifyPlan, omitting matching_target
// auto-derives it to "IP" so the policy is created and re-planned cleanly.
func TestAccFirewallPolicyResource_ipsAutoderivesMatchingTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_destinationIPsNoMatchingTarget("tf-acc-test-policy-autoderive-ip", "192.0.2.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.ips.#", "1"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.ips.0", "192.0.2.100"),
				),
			},
			// Re-running the same config must produce no drift — the bug used
			// to surface here as `~ destination.matching_target = "IP" -> "ANY"`.
			{
				Config:   testAccFirewallPolicyResourceConfig_destinationIPsNoMatchingTarget("tf-acc-test-policy-autoderive-ip", "192.0.2.100"),
				PlanOnly: true,
			},
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_networkIdAutoderivesMatchingTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyResourceConfig_destinationNetworkNoMatchingTarget("tf-acc-test-policy-autoderive-net", 3920),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.matching_target", "NETWORK"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "destination.network_id"),
				),
			},
			{
				Config:   testAccFirewallPolicyResourceConfig_destinationNetworkNoMatchingTarget("tf-acc-test-policy-autoderive-net", 3920),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFirewallPolicyResource_explicitAnyWithIpsRejected(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallPolicyResourceConfig_explicitAnyWithIPs("tf-acc-test-policy-anywithips", "192.0.2.100"),
				ExpectError: regexp.MustCompile(`(?s)matching_target="ANY" conflicts with ips`),
			},
		},
	})
}

func testAccFirewallPolicyResourceConfig_destinationIPsNoMatchingTarget(name, destIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "ALLOW"

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
    ips     = [%q]
    port    = "443"
  }
}
`, testAccProviderConfig, name, name, name, destIP)
}

func testAccFirewallPolicyResourceConfig_destinationNetworkNoMatchingTarget(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_network" "dst_net" {
  name         = "%s-dst-net"
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "ALLOW"

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id    = unifi_firewall_zone.dst.id
    network_id = unifi_network.dst_net.id
  }
}
`, testAccProviderConfig, name, name, name, vlanID, vlanID%256, vlanID%256, vlanID%256, name)
}

func testAccFirewallPolicyResourceConfig_explicitAnyWithIPs(name, destIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "ALLOW"

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id         = unifi_firewall_zone.dst.id
    matching_target = "ANY"
    ips             = [%q]
  }
}
`, testAccProviderConfig, name, name, name, destIP)
}
