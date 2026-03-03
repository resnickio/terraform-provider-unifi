package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingTeleportResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingTeleportResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_teleport.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_teleport.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_setting_teleport.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingTeleportResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingTeleportResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "enabled", "false"),
				),
			},
			{
				Config: testAccSettingTeleportResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccSettingTeleportResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_teleport" "test" {
  enabled = false
}
`
}

func testAccSettingTeleportResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_teleport" "test" {
  enabled = true
}
`
}
