package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallPolicyDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_byName("tf-acc-test-fwpolicy-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_policy.test", "name", "tf-acc-test-fwpolicy-ds"),
					resource.TestCheckResourceAttr("data.unifi_firewall_policy.test", "action", "BLOCK"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_policy.test", "id"),
				),
			},
		},
	})
}

func TestAccFirewallPolicyDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_byID("tf-acc-test-fwpolicy-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_policy.test", "name", "tf-acc-test-fwpolicy-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_policy.test", "id"),
				),
			},
		},
	})
}

func testAccFirewallPolicyDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "BLOCK"
  index  = 2099
}

data "unifi_firewall_policy" "test" {
  name = unifi_firewall_policy.test.name
}
`, testAccProviderConfig, name)
}

func testAccFirewallPolicyDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_policy" "test" {
  name   = %q
  action = "BLOCK"
  index  = 2098
}

data "unifi_firewall_policy" "test" {
  id = unifi_firewall_policy.test.id
}
`, testAccProviderConfig, name)
}
