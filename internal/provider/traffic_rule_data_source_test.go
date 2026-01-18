package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrafficRuleDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleDataSourceConfig_byName("tf-acc-test-traffic-rule-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-ds"),
					resource.TestCheckResourceAttr("data.unifi_traffic_rule.test", "action", "BLOCK"),
					resource.TestCheckResourceAttr("data.unifi_traffic_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("data.unifi_traffic_rule.test", "id"),
				),
			},
		},
	})
}

func TestAccTrafficRuleDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleDataSourceConfig_byID("tf-acc-test-traffic-rule-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_traffic_rule.test", "id"),
				),
			},
		},
	})
}

func testAccTrafficRuleDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name            = %q
  action          = "BLOCK"
  matching_target = "INTERNET"
  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

data "unifi_traffic_rule" "test" {
  name = unifi_traffic_rule.test.name
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name            = %q
  action          = "BLOCK"
  matching_target = "INTERNET"
  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

data "unifi_traffic_rule" "test" {
  id = unifi_traffic_rule.test.id
}
`, testAccProviderConfig, name)
}
