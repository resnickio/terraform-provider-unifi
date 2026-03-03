package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingGuestAccessResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_guest_access.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_guest_access.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_setting_guest_access.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingGuestAccessResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customized", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customized_title", "Welcome"),
				),
			},
		},
	})
}

func testAccSettingGuestAccessResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_guest_access" "test" {
  portal_enabled = false
}
`
}

func testAccSettingGuestAccessResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_guest_access" "test" {
  portal_enabled           = true
  portal_customized        = true
  auth                     = "none"
  portal_customized_title  = "Welcome"
}
`
}
