package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingMagicSiteToSiteVPNResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMagicSiteToSiteVPNResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_magic_site_to_site_vpn.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_magic_site_to_site_vpn.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "enabled", "false"),
				),
			},
			{
				ResourceName:            "unifi_setting_magic_site_to_site_vpn.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"x_private_key"},
			},
		},
	})
}

func TestAccSettingMagicSiteToSiteVPNResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMagicSiteToSiteVPNResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "enabled", "false"),
				),
			},
			{
				Config: testAccSettingMagicSiteToSiteVPNResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccSettingMagicSiteToSiteVPNResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_magic_site_to_site_vpn" "test" {
  enabled = false
}
`
}

func testAccSettingMagicSiteToSiteVPNResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_magic_site_to_site_vpn" "test" {
  enabled = true
}
`
}
