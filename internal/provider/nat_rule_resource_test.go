package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNatRuleResource_masquerade(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_masquerade("tf-acc-test-nat-masq"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "type", "MASQUERADE"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "description", "tf-acc-test-nat-masq"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_nat_rule.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_disabled("tf-acc-test-nat-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_logging(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_logging("tf-acc-test-nat-logging"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "logging", "true"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNatRuleResourceConfig_masquerade(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_disabled(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
  enabled     = false
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_logging(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
  logging     = true
}
`, testAccProviderConfig, description)
}
