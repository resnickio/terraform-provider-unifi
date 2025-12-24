package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

func testAccFirewallZonePreCheck(t *testing.T) {
	testAccPreCheck(t)

	config := unifi.NetworkClientConfig{
		BaseURL:            os.Getenv("UNIFI_BASE_URL"),
		APIKey:             os.Getenv("UNIFI_API_KEY"),
		Site:               "default",
		InsecureSkipVerify: os.Getenv("UNIFI_INSECURE") == "true",
	}

	client, err := unifi.NewNetworkClient(config)
	if err != nil {
		t.Skipf("Could not create client for zone support check: %v", err)
		return
	}

	_, err = client.CreateFirewallZone(context.Background(), &unifi.FirewallZone{
		Name: "tf-acc-zone-test-precheck",
	})
	if err != nil {
		t.Skipf("Controller does not support firewall zones: %v", err)
		return
	}

	zones, err := client.ListFirewallZones(context.Background())
	if err == nil {
		for _, zone := range zones {
			if zone.Name == "tf-acc-zone-test-precheck" {
				_ = client.DeleteFirewallZone(context.Background(), zone.ID)
				break
			}
		}
	}
}

func TestAccFirewallPolicyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy", "ALLOW", 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "ALLOW"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "index", "4000"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "all"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "BOTH"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "id"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create block policy
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-block", "BLOCK", 4001),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-block"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "BLOCK"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all options
			{
				Config: testAccFirewallPolicyResourceConfig_full("tf-acc-test-policy-full", 4002),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-full"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "ALLOW"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "IPV4"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "index", "4002"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "logging", "true"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "id"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-update", "ALLOW", 4003),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-update"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "action", "ALLOW"),
				),
			},
			// Update - change action and name
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-updated", "BLOCK", 4003),
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create disabled policy
			{
				Config: testAccFirewallPolicyResourceConfig_disabled("tf-acc-test-policy-disabled", 4004),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-disabled"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "enabled", "false"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create TCP policy
			{
				Config: testAccFirewallPolicyResourceConfig_protocol("tf-acc-test-policy-tcp", "tcp", 4005),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-tcp"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "tcp"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create UDP policy
			{
				Config: testAccFirewallPolicyResourceConfig_protocol("tf-acc-test-policy-udp", "udp", 4006),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-udp"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "udp"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create IPv4-only policy
			{
				Config: testAccFirewallPolicyResourceConfig_ipVersion("tf-acc-test-policy-ipv4", "IPV4", 4007),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-ipv4"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "IPV4"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create IPv6-only policy
			{
				Config: testAccFirewallPolicyResourceConfig_ipVersion("tf-acc-test-policy-ipv6", "IPV6", 4008),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-ipv6"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "IPV6"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with source
			{
				Config: testAccFirewallPolicyResourceConfig_withSource("tf-acc-test-policy-src", 4009, "10.0.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-src"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.ips.#", "1"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "source.ips.0", "10.0.0.0/24"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with destination
			{
				Config: testAccFirewallPolicyResourceConfig_withDestination("tf-acc-test-policy-dst", 4010, "192.168.1.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-dst"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.ips.#", "1"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.ips.0", "192.168.1.0/24"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with destination port
			{
				Config: testAccFirewallPolicyResourceConfig_withPort("tf-acc-test-policy-port", 4011, "443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-port"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "destination.port", "443"),
				),
			},
			// ImportState
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
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal config to verify defaults
			{
				Config: testAccFirewallPolicyResourceConfig_basic("tf-acc-test-policy-defaults", "ALLOW", 4012),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "protocol", "all"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "ip_version", "BOTH"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "logging", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "connection_state_type", "ALL"),
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "match_ipsec", "false"),
				),
			},
		},
	})
}

func TestAccFirewallPolicyResource_withZones(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccFirewallZonePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create policy with zones
			{
				Config: testAccFirewallPolicyResourceConfig_withZones("tf-acc-test-policy-zones", 4013),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_policy.test", "name", "tf-acc-test-policy-zones"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "source.zone_id"),
					resource.TestCheckResourceAttrSet("unifi_firewall_policy.test", "destination.zone_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFirewallPolicyResourceConfig_basic(name, action string, index int) string {
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
  index  = %d

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, action, index)
}

func testAccFirewallPolicyResourceConfig_full(name string, index int) string {
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
  index      = %d
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
`, testAccProviderConfig, name, name, name, index)
}

func testAccFirewallPolicyResourceConfig_disabled(name string, index int) string {
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
  index   = %d
  enabled = false

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, index)
}

func testAccFirewallPolicyResourceConfig_protocol(name, protocol string, index int) string {
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
  index    = %d
  protocol = %q

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, index, protocol)
}

func testAccFirewallPolicyResourceConfig_ipVersion(name, ipVersion string, index int) string {
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
  index      = %d
  ip_version = %q

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, index, ipVersion)
}

func testAccFirewallPolicyResourceConfig_withSource(name string, index int, sourceIP string) string {
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
  index  = %d

  source = {
    zone_id         = unifi_firewall_zone.src.id
    matching_target = "IP"
    ips             = [%q]
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
  }
}
`, testAccProviderConfig, name, name, name, index, sourceIP)
}

func testAccFirewallPolicyResourceConfig_withDestination(name string, index int, destIP string) string {
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
  index  = %d

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id         = unifi_firewall_zone.dst.id
    matching_target = "IP"
    ips             = [%q]
  }
}
`, testAccProviderConfig, name, name, name, index, destIP)
}

func testAccFirewallPolicyResourceConfig_withPort(name string, index int, port string) string {
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
  index    = %d
  protocol = "tcp"

  source = {
    zone_id = unifi_firewall_zone.src.id
  }

  destination = {
    zone_id = unifi_firewall_zone.dst.id
    port    = %q
  }
}
`, testAccProviderConfig, name, name, name, index, port)
}

func testAccFirewallPolicyResourceConfig_withZones(name string, index int) string {
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
  index  = %d

  source = {
    zone_id = unifi_firewall_zone.source.id
  }

  destination = {
    zone_id = unifi_firewall_zone.destination.id
  }
}
`, testAccProviderConfig, name, name, name, index)
}
