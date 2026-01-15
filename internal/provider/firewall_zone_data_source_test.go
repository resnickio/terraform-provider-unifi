package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallZoneDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneDataSourceConfig_byName("tf-acc-test-ds-zone"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_zone.test", "name", "tf-acc-test-ds-zone"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_zone.test", "id"),
				),
			},
		},
	})
}

func TestAccFirewallZoneDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneDataSourceConfig_byID("tf-acc-test-ds-zone-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_zone.test", "name", "tf-acc-test-ds-zone-id"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_zone.test", "id"),
				),
			},
		},
	})
}

func TestAccFirewallZoneDataSource_withNetworkIDs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneDataSourceConfig_withNetwork("tf-acc-test-ds-zone-net", "tf-acc-test-ds-zone-network", 3960),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_zone.test", "name", "tf-acc-test-ds-zone-net"),
					resource.TestCheckResourceAttr("data.unifi_firewall_zone.test", "network_ids.#", "1"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_zone.test", "id"),
				),
			},
		},
	})
}

func testAccFirewallZoneDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "source" {
  name = %q
}

data "unifi_firewall_zone" "test" {
  name = unifi_firewall_zone.source.name
}
`, testAccProviderConfig, name)
}

func testAccFirewallZoneDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "source" {
  name = %q
}

data "unifi_firewall_zone" "test" {
  id = unifi_firewall_zone.source.id
}
`, testAccProviderConfig, name)
}

func testAccFirewallZoneDataSourceConfig_withNetwork(zoneName, networkName string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = false
}

resource "unifi_firewall_zone" "source" {
  name        = %q
  network_ids = [unifi_network.test.id]
}

data "unifi_firewall_zone" "test" {
  name = unifi_firewall_zone.source.name
}
`, testAccProviderConfig, networkName, vlanID, vlanID%256, zoneName)
}
