package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingIPSResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIPSResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "site_id"),
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "dns_filtering"),
				),
			},
			{
				ResourceName:      "unifi_setting_ips.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingIPSResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIPSResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "dns_filtering"),
				),
			},
			{
				Config: testAccSettingIPSResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ips_mode", "ids"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filtering", "true"),
				),
			},
		},
	})
}

func testAccSettingIPSResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_ips" "test" {
  ips_mode = "disabled"
}
`
}

func testAccSettingIPSResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  dns_filtering = true
}
`
}
