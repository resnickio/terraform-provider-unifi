package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNatRuleDataSource_byDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleDataSourceConfig_byDescription("tf-acc-test-nat-rule-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_nat_rule.test", "description", "tf-acc-test-nat-rule-ds"),
					resource.TestCheckResourceAttr("data.unifi_nat_rule.test", "type", "MASQUERADE"),
					resource.TestCheckResourceAttrSet("data.unifi_nat_rule.test", "id"),
				),
			},
		},
	})
}

func TestAccNatRuleDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleDataSourceConfig_byID("tf-acc-test-nat-rule-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_nat_rule.test", "description", "tf-acc-test-nat-rule-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_nat_rule.test", "id"),
				),
			},
		},
	})
}

func testAccNatRuleDataSourceConfig_byDescription(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
}

data "unifi_nat_rule" "test" {
  description = unifi_nat_rule.test.description
}
`, testAccProviderConfig, description)
}

func testAccNatRuleDataSourceConfig_byID(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
}

data "unifi_nat_rule" "test" {
  id = unifi_nat_rule.test.id
}
`, testAccProviderConfig, description)
}
