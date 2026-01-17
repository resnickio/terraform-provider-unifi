package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRADIUSProfileDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRADIUSProfileDataSourceConfig_byName("tf-acc-test-radius-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_radius_profile.test", "name", "tf-acc-test-radius-ds"),
					resource.TestCheckResourceAttrSet("data.unifi_radius_profile.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_radius_profile.test", "site_id"),
					resource.TestCheckResourceAttr("data.unifi_radius_profile.test", "use_usg_auth_server", "false"),
					resource.TestCheckResourceAttr("data.unifi_radius_profile.test", "use_usg_acct_server", "false"),
					resource.TestCheckResourceAttr("data.unifi_radius_profile.test", "vlan_enabled", "false"),
					resource.TestCheckResourceAttr("data.unifi_radius_profile.test", "interim_update_enabled", "false"),
				),
			},
		},
	})
}

func TestAccRADIUSProfileDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRADIUSProfileDataSourceConfig_byID("tf-acc-test-radius-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_radius_profile.test", "name", "tf-acc-test-radius-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_radius_profile.test", "id"),
				),
			},
		},
	})
}

func testAccRADIUSProfileDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_radius_profile" "test" {
  name = %q
}

data "unifi_radius_profile" "test" {
  name = unifi_radius_profile.test.name
}
`, testAccProviderConfig, name)
}

func testAccRADIUSProfileDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_radius_profile" "test" {
  name = %q
}

data "unifi_radius_profile" "test" {
  id = unifi_radius_profile.test.id
}
`, testAccProviderConfig, name)
}
