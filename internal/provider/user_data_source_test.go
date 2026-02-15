package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_byMAC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_byMAC("aa:bb:cc:00:01:01", "tf-acc-test-user-ds-mac"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_user.test", "mac", "aa:bb:cc:00:01:01"),
					resource.TestCheckResourceAttr("data.unifi_user.test", "name", "tf-acc-test-user-ds-mac"),
					resource.TestCheckResourceAttrSet("data.unifi_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_user.test", "site_id"),
				),
			},
		},
	})
}

func TestAccUserDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_byID("aa:bb:cc:00:01:02", "tf-acc-test-user-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_user.test", "mac", "aa:bb:cc:00:01:02"),
					resource.TestCheckResourceAttr("data.unifi_user.test", "name", "tf-acc-test-user-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_user.test", "id"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig_byMAC(mac, name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user" "test" {
  mac  = %q
  name = %q
}

data "unifi_user" "test" {
  mac = unifi_user.test.mac
}
`, testAccProviderConfig, mac, name)
}

func testAccUserDataSourceConfig_byID(mac, name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user" "test" {
  mac  = %q
  name = %q
}

data "unifi_user" "test" {
  id = unifi_user.test.id
}
`, testAccProviderConfig, mac, name)
}
