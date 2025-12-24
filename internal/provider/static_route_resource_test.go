package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStaticRouteResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccStaticRouteResourceConfig_basic("tf-acc-test-route", "10.99.0.0/24", "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_network", "10.99.0.0/24"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_nexthop", "192.168.1.1"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "type", "static-route"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_type", "nexthop-route"),
					resource.TestCheckResourceAttrSet("unifi_static_route.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_static_route.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_static_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticRouteResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all options
			{
				Config: testAccStaticRouteResourceConfig_full("tf-acc-test-route-full", "10.98.0.0/24", "192.168.1.1", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-full"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_network", "10.98.0.0/24"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_nexthop", "192.168.1.1"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_distance", "10"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_static_route.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_static_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticRouteResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccStaticRouteResourceConfig_basic("tf-acc-test-route-update", "10.97.0.0/24", "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-update"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_nexthop", "192.168.1.1"),
				),
			},
			// Update - change nexthop
			{
				Config: testAccStaticRouteResourceConfig_basic("tf-acc-test-route-updated", "10.97.0.0/24", "192.168.1.254"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-updated"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_nexthop", "192.168.1.254"),
				),
			},
		},
	})
}

func TestAccStaticRouteResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create disabled route
			{
				Config: testAccStaticRouteResourceConfig_disabled("tf-acc-test-route-disabled", "10.96.0.0/24", "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-disabled"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "enabled", "false"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_static_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticRouteResource_blackhole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create blackhole route
			{
				Config: testAccStaticRouteResourceConfig_blackhole("tf-acc-test-route-blackhole", "10.95.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-blackhole"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_type", "blackhole"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_static_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticRouteResource_defaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal config to verify defaults
			{
				Config: testAccStaticRouteResourceConfig_minimal("tf-acc-test-route-defaults", "10.94.0.0/24", "192.168.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-defaults"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "type", "static-route"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_type", "nexthop-route"),
				),
			},
		},
	})
}

func TestAccStaticRouteResource_distance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with administrative distance
			{
				Config: testAccStaticRouteResourceConfig_full("tf-acc-test-route-distance", "10.93.0.0/24", "192.168.1.1", 50),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "name", "tf-acc-test-route-distance"),
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_distance", "50"),
				),
			},
			// Update - change distance
			{
				Config: testAccStaticRouteResourceConfig_full("tf-acc-test-route-distance", "10.93.0.0/24", "192.168.1.1", 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_route.test", "static_route_distance", "100"),
				),
			},
		},
	})
}

func testAccStaticRouteResourceConfig_basic(name, network, nexthop string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                 = %q
  static_route_network = %q
  static_route_nexthop = %q
}
`, testAccProviderConfig, name, network, nexthop)
}

func testAccStaticRouteResourceConfig_full(name, network, nexthop string, distance int) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                   = %q
  static_route_network   = %q
  static_route_nexthop   = %q
  static_route_distance  = %d
  enabled                = true
}
`, testAccProviderConfig, name, network, nexthop, distance)
}

func testAccStaticRouteResourceConfig_disabled(name, network, nexthop string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                 = %q
  static_route_network = %q
  static_route_nexthop = %q
  enabled              = false
}
`, testAccProviderConfig, name, network, nexthop)
}

func testAccStaticRouteResourceConfig_blackhole(name, network string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                 = %q
  static_route_network = %q
  static_route_type    = "blackhole"
}
`, testAccProviderConfig, name, network)
}

func testAccStaticRouteResourceConfig_minimal(name, network, nexthop string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_route" "test" {
  name                  = %q
  static_route_network  = %q
  static_route_nexthop  = %q
}
`, testAccProviderConfig, name, network, nexthop)
}
