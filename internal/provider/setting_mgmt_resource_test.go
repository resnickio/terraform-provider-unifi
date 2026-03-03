package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingMgmtResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_mgmt.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_mgmt.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "auto_upgrade", "false"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "led_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "alert_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "x_ssh_enabled", "false"),
				),
			},
			{
				ResourceName:            "unifi_setting_mgmt.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"x_ssh_password"},
			},
		},
	})
}

func TestAccSettingMgmtResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "auto_upgrade", "false"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "led_enabled", "true"),
				),
			},
			{
				Config: testAccSettingMgmtResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "auto_upgrade", "true"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "led_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "x_ssh_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "x_ssh_auth_password_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_mgmt.test", "x_ssh_username", "admin"),
				),
			},
		},
	})
}

func testAccSettingMgmtResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_mgmt" "test" {
  auto_upgrade = false
  led_enabled  = true
  alert_enabled = true
}
`
}

func testAccSettingMgmtResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_mgmt" "test" {
  auto_upgrade                = true
  led_enabled                 = false
  x_ssh_enabled               = true
  x_ssh_auth_password_enabled = true
  x_ssh_username              = "admin"
  x_ssh_password              = "testpass123"
}
`
}
