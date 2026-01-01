package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallZoneResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccFirewallZoneResourceConfig_basic("tf-acc-test-zone"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "name", "tf-acc-test-zone"),
					resource.TestCheckResourceAttrSet("unifi_firewall_zone.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallZoneResource_readBuiltInZone(t *testing.T) {
	t.Skip("Import-only tests for built-in zones not supported by test framework")
}

func TestAccFirewallZoneResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccFirewallZoneResourceConfig_basic("tf-acc-test-zone-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "name", "tf-acc-test-zone-update"),
				),
			},
			// Update - change name
			{
				Config: testAccFirewallZoneResourceConfig_basic("tf-acc-test-zone-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "name", "tf-acc-test-zone-updated"),
				),
			},
		},
	})
}

func TestAccFirewallZoneResource_withNetworkIDs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create zone and network, assign network to zone
			{
				Config: testAccFirewallZoneResourceConfig_withNetwork("tf-acc-test-zone-networks", "tf-acc-test-network-zone", 3950),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "name", "tf-acc-test-zone-networks"),
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "network_ids.#", "1"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallZoneResource_multipleNetworks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCheckControllerSupportsZones(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create zone with multiple networks
			{
				Config: testAccFirewallZoneResourceConfig_multipleNetworks("tf-acc-test-zone-multi", 3951, 3952),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "name", "tf-acc-test-zone-multi"),
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "network_ids.#", "2"),
				),
			},
			// Update - remove one network
			{
				Config: testAccFirewallZoneResourceConfig_withNetwork("tf-acc-test-zone-multi", "tf-acc-test-network-zone-1", 3951),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "network_ids.#", "1"),
				),
			},
		},
	})
}

func testAccFirewallZoneResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_zone" "test" {
  name = %q
}
`, testAccProviderConfig, name)
}

func testAccFirewallZoneResourceConfig_withNetwork(zoneName, networkName string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = false
}

resource "unifi_firewall_zone" "test" {
  name        = %q
  network_ids = [unifi_network.test.id]
}
`, testAccProviderConfig, networkName, vlanID, vlanID%256, zoneName)
}

func testAccFirewallZoneResourceConfig_multipleNetworks(zoneName string, vlanID1, vlanID2 int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test1" {
  name         = "tf-acc-test-network-zone-1"
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = false
}

resource "unifi_network" "test2" {
  name         = "tf-acc-test-network-zone-2"
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = false
}

resource "unifi_firewall_zone" "test" {
  name        = %q
  network_ids = [unifi_network.test1.id, unifi_network.test2.id]
}
`, testAccProviderConfig, vlanID1, vlanID1%256, vlanID2, vlanID2%256, zoneName)
}
