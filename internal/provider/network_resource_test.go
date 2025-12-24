package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccNetworkResourceConfig_basic("tf-acc-test-network", 3900),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network"),
					resource.TestCheckResourceAttr("unifi_network.test", "purpose", "corporate"),
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", "3900"),
					resource.TestCheckResourceAttrSet("unifi_network.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_network.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all options
			{
				Config: testAccNetworkResourceConfig_full("tf-acc-test-network-full", 3901),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-full"),
					resource.TestCheckResourceAttr("unifi_network.test", "purpose", "corporate"),
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", "3901"),
					resource.TestCheckResourceAttr("unifi_network.test", "subnet", "10.39.1.1/24"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_start", "10.39.1.10"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_stop", "10.39.1.254"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_lease", "43200"),
					resource.TestCheckResourceAttr("unifi_network.test", "domain_name", "test.local"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "2"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.1", "8.8.4.4"),
					resource.TestCheckResourceAttr("unifi_network.test", "igmp_snooping", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_network.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccNetworkResourceConfig_basic("tf-acc-test-network-update", 3902),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-update"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_enabled", "true"),
				),
			},
			// Update - change name and disable DHCP
			{
				Config: testAccNetworkResourceConfig_updated("tf-acc-test-network-updated", 3902),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-updated"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_enabled", "false"),
				),
			},
		},
	})
}

func TestAccNetworkResource_defaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal config to verify defaults
			{
				Config: testAccNetworkResourceConfig_minimal("tf-acc-test-network-defaults", 3903),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-defaults"),
					// Verify defaults are applied
					resource.TestCheckResourceAttr("unifi_network.test", "network_group", "LAN"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_lease", "86400"),
					resource.TestCheckResourceAttr("unifi_network.test", "igmp_snooping", "false"),
					resource.TestCheckResourceAttr("unifi_network.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccNetworkResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create disabled network
			{
				Config: testAccNetworkResourceConfig_disabled("tf-acc-test-network-disabled", 3904),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-disabled"),
					resource.TestCheckResourceAttr("unifi_network.test", "enabled", "false"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_igmpSnooping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with IGMP snooping enabled
			{
				Config: testAccNetworkResourceConfig_igmpSnooping("tf-acc-test-network-igmp", 3905),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-igmp"),
					resource.TestCheckResourceAttr("unifi_network.test", "igmp_snooping", "true"),
				),
			},
			// Update - disable IGMP snooping
			{
				Config: testAccNetworkResourceConfig_igmpSnoopingDisabled("tf-acc-test-network-igmp", 3905),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "igmp_snooping", "false"),
				),
			},
		},
	})
}

func TestAccNetworkResource_guest(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create guest network
			{
				Config: testAccNetworkResourceConfig_guest("tf-acc-test-network-guest", 3906),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-guest"),
					resource.TestCheckResourceAttr("unifi_network.test", "purpose", "guest"),
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", "3906"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_vlanOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create VLAN-only network (no DHCP/routing)
			{
				Config: testAccNetworkResourceConfig_vlanOnly("tf-acc-test-network-vlan-only", 3907),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-vlan-only"),
					resource.TestCheckResourceAttr("unifi_network.test", "purpose", "vlan-only"),
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", "3907"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_dhcpDnsServers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with 1 DNS server
			{
				Config: testAccNetworkResourceConfig_dhcpDns("tf-acc-test-network-dns", 3908, []string{"1.1.1.1"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "1"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.0", "1.1.1.1"),
				),
			},
			// Update to 4 DNS servers (max)
			{
				Config: testAccNetworkResourceConfig_dhcpDns("tf-acc-test-network-dns", 3908, []string{"1.1.1.1", "8.8.8.8", "8.8.4.4", "9.9.9.9"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "4"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.0", "1.1.1.1"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.1", "8.8.8.8"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.2", "8.8.4.4"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.3", "9.9.9.9"),
				),
			},
		},
	})
}

func testAccNetworkResourceConfig_basic(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_full(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name          = %q
  purpose       = "corporate"
  vlan_id       = %d
  subnet        = "10.39.1.1/24"
  dhcp_enabled  = true
  dhcp_start    = "10.39.1.10"
  dhcp_stop     = "10.39.1.254"
  dhcp_lease    = 43200
  dhcp_dns      = ["8.8.8.8", "8.8.4.4"]
  domain_name   = "test.local"
  igmp_snooping = true
  enabled       = true
}
`, testAccProviderConfig, name, vlanID)
}

func testAccNetworkResourceConfig_updated(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = false
}
`, testAccProviderConfig, name, vlanID, vlanID%256)
}

func testAccNetworkResourceConfig_minimal(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name    = %q
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}
`, testAccProviderConfig, name, vlanID, vlanID%256)
}

func testAccNetworkResourceConfig_disabled(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = false
  enabled      = false
}
`, testAccProviderConfig, name, vlanID, vlanID%256)
}

func testAccNetworkResourceConfig_igmpSnooping(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name          = %q
  purpose       = "corporate"
  vlan_id       = %d
  subnet        = "10.%d.0.1/24"
  dhcp_enabled  = true
  dhcp_start    = "10.%d.0.10"
  dhcp_stop     = "10.%d.0.254"
  igmp_snooping = true
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_igmpSnoopingDisabled(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name          = %q
  purpose       = "corporate"
  vlan_id       = %d
  subnet        = "10.%d.0.1/24"
  dhcp_enabled  = true
  dhcp_start    = "10.%d.0.10"
  dhcp_stop     = "10.%d.0.254"
  igmp_snooping = false
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_guest(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "guest"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_vlanOnly(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name    = %q
  purpose = "vlan-only"
  vlan_id = %d
}
`, testAccProviderConfig, name, vlanID)
}

func testAccNetworkResourceConfig_dhcpDns(name string, vlanID int, dnsServers []string) string {
	dnsStr := ""
	for i, dns := range dnsServers {
		if i > 0 {
			dnsStr += ", "
		}
		dnsStr += fmt.Sprintf("%q", dns)
	}

	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
  dhcp_dns     = [%s]
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, dnsStr)
}
