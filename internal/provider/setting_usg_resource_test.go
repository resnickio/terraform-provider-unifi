package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingUSGResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUSGResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_usg.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "broadcast_ping", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "mdns_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_setting_usg.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingUSGResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUSGResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_enabled", "false"),
				),
			},
			{
				Config: testAccSettingUSGResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "broadcast_ping", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_nat_pmp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp_secure_mode", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "lldp_enable_all", "true"),
				),
			},
		},
	})
}

func testAccSettingUSGResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_usg" "test" {
  broadcast_ping = false
  mdns_enabled   = false
  upnp_enabled   = false
}
`
}

func testAccSettingUSGResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_usg" "test" {
  broadcast_ping       = true
  upnp_enabled         = true
  upnp_nat_pmp_enabled = true
  upnp_secure_mode     = true
  lldp_enable_all      = true
}
`
}
