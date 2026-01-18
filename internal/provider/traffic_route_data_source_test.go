package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrafficRouteDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteDataSourceConfig_byName("tf-acc-test-traffic-route-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-ds"),
					resource.TestCheckResourceAttr("data.unifi_traffic_route.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("data.unifi_traffic_route.test", "id"),
				),
			},
		},
	})
}

func TestAccTrafficRouteDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteDataSourceConfig_byID("tf-acc-test-traffic-route-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_traffic_route.test", "id"),
				),
			},
		},
	})
}

func testAccTrafficRouteDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name           = %q
  matching_target = "INTERNET"
  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

data "unifi_traffic_route" "test" {
  name = unifi_traffic_route.test.name
}
`, testAccProviderConfig, name)
}

func testAccTrafficRouteDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name           = %q
  matching_target = "INTERNET"
  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

data "unifi_traffic_route" "test" {
  id = unifi_traffic_route.test.id
}
`, testAccProviderConfig, name)
}
