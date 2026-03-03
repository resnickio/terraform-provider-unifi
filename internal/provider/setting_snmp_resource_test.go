package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingSNMPResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingSNMPResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_snmp.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_setting_snmp.test", "site_id"),
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "enabled_v3", "false"),
				),
			},
			{
				ResourceName:            "unifi_setting_snmp.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"x_password"},
			},
		},
	})
}

func TestAccSettingSNMPResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingSNMPResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "enabled", "false"),
				),
			},
			{
				Config: testAccSettingSNMPResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "community", "public"),
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "enabled_v3", "true"),
					resource.TestCheckResourceAttr("unifi_setting_snmp.test", "username", "snmpuser"),
				),
			},
		},
	})
}

func testAccSettingSNMPResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_setting_snmp" "test" {
  enabled = false
}
`
}

func testAccSettingSNMPResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_setting_snmp" "test" {
  enabled    = true
  community  = "public"
  enabled_v3 = true
  username   = "snmpuser"
  x_password = "snmppass123"
}
`
}
