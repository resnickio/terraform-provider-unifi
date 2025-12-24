package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallRuleResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccFirewallRuleResourceConfig_basic("tf-acc-test-rule", 4001),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "ruleset", "LAN_IN"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "rule_index", "4001"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_port", "22"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_withGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rule with firewall groups
			{
				Config: testAccFirewallRuleResourceConfig_withDstGroups("tf-acc-test-rule-groups", 4002),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-groups"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "accept"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_firewall_group_ids.#", "1"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_withSrcGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rule with source firewall groups
			{
				Config: testAccFirewallRuleResourceConfig_withSrcGroups("tf-acc-test-rule-srcgroups", 4003),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-srcgroups"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_firewall_group_ids.#", "1"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccFirewallRuleResourceConfig_basic("tf-acc-test-rule-update", 4004),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-update"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "enabled", "true"),
				),
			},
			// Update - change action and disable
			{
				Config: testAccFirewallRuleResourceConfig_updated("tf-acc-test-rule-update", 4004),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-update"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "reject"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "logging", "true"),
				),
			},
		},
	})
}

func TestAccFirewallRuleResource_allProtocol(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rule with protocol "all" (no dst_port)
			{
				Config: testAccFirewallRuleResourceConfig_allProtocol("tf-acc-test-rule-all", 4005),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-all"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "all"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_defaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal config to verify defaults
			{
				Config: testAccFirewallRuleResourceConfig_minimal("tf-acc-test-rule-defaults", 4006),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-defaults"),
					// Verify defaults are applied
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "all"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "logging", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_new", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_established", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_related", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_invalid", "false"),
				),
			},
		},
	})
}

func TestAccFirewallRuleResource_withSrcAddress(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rule with source address
			{
				Config: testAccFirewallRuleResourceConfig_withSrcAddress("tf-acc-test-rule-srcaddr", 4007, "192.168.1.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-srcaddr"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_address", "192.168.1.100"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_network_conf_type", "ADDRv4"),
				),
			},
			// Update - change to CIDR
			{
				Config: testAccFirewallRuleResourceConfig_withSrcNetwork("tf-acc-test-rule-srcaddr", 4007, "192.168.1.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_address", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_network_conf_type", "NETv4"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_withDstAddress(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rule with destination address
			{
				Config: testAccFirewallRuleResourceConfig_withDstAddress("tf-acc-test-rule-dstaddr", 4008, "10.0.0.50"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-dstaddr"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_address", "10.0.0.50"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_network_conf_type", "ADDRv4"),
				),
			},
			// Update - change to network
			{
				Config: testAccFirewallRuleResourceConfig_withDstNetwork("tf-acc-test-rule-dstaddr", 4008, "10.0.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_address", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_network_conf_type", "NETv4"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_stateTracking(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with custom state tracking
			{
				Config: testAccFirewallRuleResourceConfig_stateTracking("tf-acc-test-rule-state", 4009),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-state"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_new", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_established", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_related", "false"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_invalid", "true"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_withLogging(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with logging enabled
			{
				Config: testAccFirewallRuleResourceConfig_withLogging("tf-acc-test-rule-log", 4010),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-log"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "logging", "true"),
				),
			},
			// Update - disable logging
			{
				Config: testAccFirewallRuleResourceConfig_basic("tf-acc-test-rule-log", 4010),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "logging", "false"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_udpProtocol(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create UDP rule
			{
				Config: testAccFirewallRuleResourceConfig_udp("tf-acc-test-rule-udp", 4011),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-udp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "udp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_port", "53"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_tcpUdpProtocol(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create TCP+UDP rule
			{
				Config: testAccFirewallRuleResourceConfig_tcpUdp("tf-acc-test-rule-tcpudp", 4012),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-tcpudp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "tcp_udp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_port", "53"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_icmpProtocol(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create ICMP rule (no port)
			{
				Config: testAccFirewallRuleResourceConfig_icmp("tf-acc-test-rule-icmp", 4013),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-icmp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "icmp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "accept"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_wanRuleset(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create WAN_IN rule
			{
				Config: testAccFirewallRuleResourceConfig_wanIn("tf-acc-test-rule-wanin", 4014),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-wanin"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "ruleset", "WAN_IN"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_wanLocalRuleset(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create WAN_LOCAL rule
			{
				Config: testAccFirewallRuleResourceConfig_wanLocal("tf-acc-test-rule-wanlocal", 4015),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-wanlocal"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "ruleset", "WAN_LOCAL"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_acceptAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create accept rule
			{
				Config: testAccFirewallRuleResourceConfig_accept("tf-acc-test-rule-accept", 4016),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-accept"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "accept"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_rejectAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create reject rule
			{
				Config: testAccFirewallRuleResourceConfig_reject("tf-acc-test-rule-reject", 4017),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-reject"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "reject"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_portRange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create rule with port range
			{
				Config: testAccFirewallRuleResourceConfig_portRange("tf-acc-test-rule-portrange", 4018, "8080-8090"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-portrange"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_port", "8080-8090"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallRuleResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all options
			{
				Config: testAccFirewallRuleResourceConfig_full("tf-acc-test-rule-full", 4019),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "name", "tf-acc-test-rule-full"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "ruleset", "LAN_IN"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "action", "drop"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "rule_index", "4019"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_network_conf_type", "NETv4"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "src_address", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_network_conf_type", "ADDRv4"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_address", "10.0.0.1"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "dst_port", "443"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "logging", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_new", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_established", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_related", "true"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.test", "state_invalid", "false"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_firewall_rule.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFirewallRuleResourceConfig_basic(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = %d
  enabled    = true
  protocol   = "tcp"
  dst_port   = "22"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_withDstGroups(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "test_ports" {
  name       = "tf-acc-test-ports"
  group_type = "port-group"
  members    = ["80", "443"]
}

resource "unifi_firewall_rule" "test" {
  name                   = %q
  ruleset                = "LAN_IN"
  action                 = "accept"
  rule_index             = %d
  enabled                = true
  protocol               = "tcp"
  dst_firewall_group_ids = [unifi_firewall_group.test_ports.id]
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_withSrcGroups(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "test_addrs" {
  name       = "tf-acc-test-addrs"
  group_type = "address-group"
  members    = ["192.168.1.100", "192.168.1.101"]
}

resource "unifi_firewall_rule" "test" {
  name                   = %q
  ruleset                = "LAN_IN"
  action                 = "drop"
  rule_index             = %d
  enabled                = true
  protocol               = "all"
  src_firewall_group_ids = [unifi_firewall_group.test_addrs.id]
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_updated(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "reject"
  rule_index = %d
  enabled    = false
  protocol   = "tcp"
  dst_port   = "22"
  logging    = true
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_allProtocol(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "WAN_IN"
  action     = "drop"
  rule_index = %d
  enabled    = true
  protocol   = "all"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_minimal(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = %d
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_withSrcAddress(name string, ruleIndex int, srcAddr string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name                 = %q
  ruleset              = "LAN_IN"
  action               = "drop"
  rule_index           = %d
  enabled              = true
  protocol             = "all"
  src_network_conf_type = "ADDRv4"
  src_address          = %q
}
`, testAccProviderConfig, name, ruleIndex, srcAddr)
}

func testAccFirewallRuleResourceConfig_withSrcNetwork(name string, ruleIndex int, srcNet string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name                 = %q
  ruleset              = "LAN_IN"
  action               = "drop"
  rule_index           = %d
  enabled              = true
  protocol             = "all"
  src_network_conf_type = "NETv4"
  src_address          = %q
}
`, testAccProviderConfig, name, ruleIndex, srcNet)
}

func testAccFirewallRuleResourceConfig_withDstAddress(name string, ruleIndex int, dstAddr string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name                 = %q
  ruleset              = "LAN_IN"
  action               = "drop"
  rule_index           = %d
  enabled              = true
  protocol             = "tcp"
  dst_network_conf_type = "ADDRv4"
  dst_address          = %q
  dst_port             = "443"
}
`, testAccProviderConfig, name, ruleIndex, dstAddr)
}

func testAccFirewallRuleResourceConfig_withDstNetwork(name string, ruleIndex int, dstNet string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name                 = %q
  ruleset              = "LAN_IN"
  action               = "drop"
  rule_index           = %d
  enabled              = true
  protocol             = "tcp"
  dst_network_conf_type = "NETv4"
  dst_address          = %q
  dst_port             = "443"
}
`, testAccProviderConfig, name, ruleIndex, dstNet)
}

func testAccFirewallRuleResourceConfig_stateTracking(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name              = %q
  ruleset           = "LAN_IN"
  action            = "drop"
  rule_index        = %d
  enabled           = true
  protocol          = "all"
  state_new         = true
  state_established = false
  state_related     = false
  state_invalid     = true
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_withLogging(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = %d
  enabled    = true
  protocol   = "tcp"
  dst_port   = "22"
  logging    = true
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_udp(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "accept"
  rule_index = %d
  enabled    = true
  protocol   = "udp"
  dst_port   = "53"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_tcpUdp(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "accept"
  rule_index = %d
  enabled    = true
  protocol   = "tcp_udp"
  dst_port   = "53"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_icmp(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "accept"
  rule_index = %d
  enabled    = true
  protocol   = "icmp"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_wanIn(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "WAN_IN"
  action     = "drop"
  rule_index = %d
  enabled    = true
  protocol   = "all"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_wanLocal(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "WAN_LOCAL"
  action     = "drop"
  rule_index = %d
  enabled    = true
  protocol   = "tcp"
  dst_port   = "22"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_accept(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "accept"
  rule_index = %d
  enabled    = true
  protocol   = "tcp"
  dst_port   = "443"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_reject(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "reject"
  rule_index = %d
  enabled    = true
  protocol   = "tcp"
  dst_port   = "23"
}
`, testAccProviderConfig, name, ruleIndex)
}

func testAccFirewallRuleResourceConfig_portRange(name string, ruleIndex int, portRange string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = %d
  enabled    = true
  protocol   = "tcp"
  dst_port   = %q
}
`, testAccProviderConfig, name, ruleIndex, portRange)
}

func testAccFirewallRuleResourceConfig_full(name string, ruleIndex int) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name                  = %q
  ruleset               = "LAN_IN"
  action                = "drop"
  rule_index            = %d
  enabled               = true
  protocol              = "tcp"
  src_network_conf_type = "NETv4"
  src_address           = "192.168.1.0/24"
  dst_network_conf_type = "ADDRv4"
  dst_address           = "10.0.0.1"
  dst_port              = "443"
  logging               = true
  state_new             = true
  state_established     = true
  state_related         = true
  state_invalid         = false
}
`, testAccProviderConfig, name, ruleIndex)
}
