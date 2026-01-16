package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortProfileDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileDataSourceConfig_byName("tf-acc-test-port-profile-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_port_profile.test", "name", "tf-acc-test-port-profile-ds"),
					resource.TestCheckResourceAttrSet("data.unifi_port_profile.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_port_profile.test", "site_id"),
				),
			},
		},
	})
}

func TestAccPortProfileDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileDataSourceConfig_byID("tf-acc-test-port-profile-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_port_profile.test", "name", "tf-acc-test-port-profile-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_port_profile.test", "id"),
				),
			},
		},
	})
}

func testAccPortProfileDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_profile" "test" {
  name = %q
}

data "unifi_port_profile" "test" {
  name = unifi_port_profile.test.name
}
`, testAccProviderConfig, name)
}

func testAccPortProfileDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_profile" "test" {
  name = %q
}

data "unifi_port_profile" "test" {
  id = unifi_port_profile.test.id
}
`, testAccProviderConfig, name)
}
