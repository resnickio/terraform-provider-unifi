package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStaticRouteDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticRouteDataSourceConfig_byName("tf-acc-test-route-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_static_route.test", "name", "tf-acc-test-route-ds"),
					resource.TestCheckResourceAttr("data.unifi_static_route.test", "static_route_network", "192.168.200.0/24"),
					resource.TestCheckResourceAttrSet("data.unifi_static_route.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_static_route.test", "site_id"),
				),
			},
		},
	})
}

func TestAccStaticRouteDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticRouteDataSourceConfig_byID("tf-acc-test-route-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_static_route.test", "name", "tf-acc-test-route-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_static_route.test", "id"),
				),
			},
		},
	})
}

func testAccStaticRouteDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                 = %q
  static_route_network = "192.168.200.0/24"
  static_route_type    = "blackhole"
}

data "unifi_static_route" "test" {
  name = unifi_static_route.test.name
}
`, testAccProviderConfig, name)
}

func testAccStaticRouteDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                 = %q
  static_route_network = "192.168.201.0/24"
  static_route_type    = "blackhole"
}

data "unifi_static_route" "test" {
  id = unifi_static_route.test.id
}
`, testAccProviderConfig, name)
}
