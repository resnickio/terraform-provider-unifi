package provider

import (
	"fmt"
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
