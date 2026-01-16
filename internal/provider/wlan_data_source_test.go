package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWLANDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANDataSourceConfig_byName("tf-acc-test-wlan-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_wlan.test", "name", "tf-acc-test-wlan-ds"),
					resource.TestCheckResourceAttr("data.unifi_wlan.test", "security", "wpapsk"),
					resource.TestCheckResourceAttrSet("data.unifi_wlan.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_wlan.test", "site_id"),
				),
			},
		},
	})
}

func TestAccWLANDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANDataSourceConfig_byID("tf-acc-test-wlan-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_wlan.test", "name", "tf-acc-test-wlan-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_wlan.test", "id"),
				),
			},
		},
	})
}

func testAccWLANDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

data "unifi_ap_group" "default" {
  name = "Default"
}

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = "testpassword123"
  ap_group_ids = [data.unifi_ap_group.default.id]
}

data "unifi_wlan" "test" {
  name = unifi_wlan.test.name
}
`, testAccProviderConfig, name)
}

func testAccWLANDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

data "unifi_ap_group" "default" {
  name = "Default"
}

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = "testpassword123"
  ap_group_ids = [data.unifi_ap_group.default.id]
}

data "unifi_wlan" "test" {
  id = unifi_wlan.test.id
}
`, testAccProviderConfig, name)
}
