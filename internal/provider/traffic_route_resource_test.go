package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrafficRouteResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteResourceConfig_basic("tf-acc-test-traffic-route-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-basic"),
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_traffic_route.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_traffic_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRouteResource_withDomains(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteResourceConfig_withDomains("tf-acc-test-traffic-route-domains"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-domains"),
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "matching_target", "DOMAIN"),
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "domains.#", "2"),
				),
			},
			{
				ResourceName:      "unifi_traffic_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRouteResource_withIPAddresses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteResourceConfig_withIPAddresses("tf-acc-test-traffic-route-ips"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-ips"),
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "ip_addresses.#", "2"),
				),
			},
			{
				ResourceName:      "unifi_traffic_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRouteResource_killSwitch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteResourceConfig_killSwitch("tf-acc-test-traffic-route-ks"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-ks"),
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "kill_switch", "true"),
				),
			},
			{
				ResourceName:      "unifi_traffic_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRouteResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteResourceConfig_disabled("tf-acc-test-traffic-route-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_traffic_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRouteResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRouteResourceConfig_basic("tf-acc-test-traffic-route-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-update"),
				),
			},
			{
				Config: testAccTrafficRouteResourceConfig_basic("tf-acc-test-traffic-route-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_route.test", "name", "tf-acc-test-traffic-route-updated"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_traffic_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTrafficRouteResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name        = %q
  description = "Test traffic route"
}
`, testAccProviderConfig, name)
}

func testAccTrafficRouteResourceConfig_withDomains(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name            = %q
  matching_target = "DOMAIN"

  domains = [
    {
      domain      = "*.example.com"
      description = "Example domain"
    },
    {
      domain = "test.local"
    }
  ]
}
`, testAccProviderConfig, name)
}

func testAccTrafficRouteResourceConfig_withIPAddresses(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name            = %q
  matching_target = "IP"
  ip_addresses    = ["192.168.1.0/24", "10.0.0.0/8"]
}
`, testAccProviderConfig, name)
}

func testAccTrafficRouteResourceConfig_killSwitch(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name        = %q
  kill_switch = true
}
`, testAccProviderConfig, name)
}

func testAccTrafficRouteResourceConfig_disabled(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_route" "test" {
  name    = %q
  enabled = false
}
`, testAccProviderConfig, name)
}
