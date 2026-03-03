package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingRadiusResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRadiusResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_radius.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_radius.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "accounting_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "auth_port", "1812"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "acct_port", "1813"),
				),
			},
			{
				ResourceName:            "unifi_setting_radius.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"x_secret"},
			},
		},
	})
}

func TestAccSettingRadiusResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRadiusResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "enabled", "false"),
				),
			},
			{
				Config: testAccSettingRadiusResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "accounting_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "auth_port", "1812"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "acct_port", "1813"),
					resource.TestCheckResourceAttr("unifi_setting_radius.test", "tunneled_reply", "true"),
				),
			},
		},
	})
}

func testAccSettingRadiusResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_radius" "test" {
  enabled            = false
  accounting_enabled = false
}
`
}

func testAccSettingRadiusResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_radius" "test" {
  enabled            = true
  accounting_enabled = true
  auth_port          = 1812
  acct_port          = 1813
  x_secret           = "testsecret123"
  tunneled_reply     = true
}
`
}
