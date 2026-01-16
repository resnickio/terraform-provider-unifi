package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallRuleDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleDataSourceConfig_byName("tf-acc-test-fwrule-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_rule.test", "name", "tf-acc-test-fwrule-ds"),
					resource.TestCheckResourceAttr("data.unifi_firewall_rule.test", "ruleset", "WAN_IN"),
					resource.TestCheckResourceAttr("data.unifi_firewall_rule.test", "action", "drop"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_rule.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_rule.test", "site_id"),
				),
			},
		},
	})
}

func TestAccFirewallRuleDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleDataSourceConfig_byID("tf-acc-test-fwrule-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_rule.test", "name", "tf-acc-test-fwrule-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_rule.test", "id"),
				),
			},
		},
	})
}

func testAccFirewallRuleDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "WAN_IN"
  action     = "drop"
  rule_index = 2099
  protocol   = "all"
}

data "unifi_firewall_rule" "test" {
  name = unifi_firewall_rule.test.name
}
`, testAccProviderConfig, name)
}

func testAccFirewallRuleDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_rule" "test" {
  name       = %q
  ruleset    = "WAN_IN"
  action     = "drop"
  rule_index = 2098
  protocol   = "all"
}

data "unifi_firewall_rule" "test" {
  id = unifi_firewall_rule.test.id
}
`, testAccProviderConfig, name)
}
