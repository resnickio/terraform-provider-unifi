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
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "8.8.8.8"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "8.8.4.4"),
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
					// Network access - these are set by API, just verify they exist
					resource.TestCheckResourceAttr("unifi_network.test", "internet_access_enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_network.test", "intra_network_access_enabled"),
					resource.TestCheckResourceAttr("unifi_network.test", "nat_enabled", "true"),
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
		PreCheck:                 func() { testAccCheckControllerSupportsGuestNetworks(t) },
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
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "1.1.1.1"),
				),
			},
			// Update to 4 DNS servers (max)
			{
				Config: testAccNetworkResourceConfig_dhcpDns("tf-acc-test-network-dns", 3908, []string{"1.1.1.1", "8.8.8.8", "8.8.4.4", "9.9.9.9"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "4"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "1.1.1.1"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "8.8.8.8"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "8.8.4.4"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "9.9.9.9"),
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
  name    = %q
  purpose = "guest"
  vlan_id = %d
}
`, testAccProviderConfig, name, vlanID)
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
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, formatStringListForHCL(dnsServers))
}

func TestAccNetworkResource_dhcpNtp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_dhcpNtp("tf-acc-test-network-ntp", 3909, []string{"129.6.15.28"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-ntp"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp.#", "1"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_ntp.*", "129.6.15.28"),
				),
			},
			{
				Config: testAccNetworkResourceConfig_dhcpNtp("tf-acc-test-network-ntp", 3909, []string{"129.6.15.28", "129.6.15.29"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_ntp.*", "129.6.15.28"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_ntp.*", "129.6.15.29"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_dhcpGateway(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_dhcpGateway("tf-acc-test-network-gw", 3910),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-gw"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_gateway_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_gateway", "10.70.0.254"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_dhcpBoot(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_dhcpBoot("tf-acc-test-network-pxe", 3911),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-pxe"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_server", "10.71.0.10"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_filename", "pxelinux.0"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_dhcpAdvanced(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_dhcpAdvanced("tf-acc-test-network-adv", 3912),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-adv"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_unifi_controller", "10.72.0.1"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_wpad_url", "http://wpad.test.local/wpad.dat"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_time_offset_enabled", "true"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_networkAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_networkAccess("tf-acc-test-network-access", 3913),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-access"),
					resource.TestCheckResourceAttr("unifi_network.test", "internet_access_enabled", "false"),
					// intra_network_access_enabled may not be supported on all controller types
					resource.TestCheckResourceAttrSet("unifi_network.test", "intra_network_access_enabled"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_mdns(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_mdns("tf-acc-test-network-mdns", 3914),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-mdns"),
					resource.TestCheckResourceAttr("unifi_network.test", "mdns_enabled", "true"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkResource_allDhcpOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_allDhcpOptions("tf-acc-test-network-all-dhcp", 3915),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-all-dhcp"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "2"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp.#", "1"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_gateway_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_gateway", "10.75.0.254"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_server", "10.75.0.10"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_filename", "pxelinux.0"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_unifi_controller", "10.75.0.1"),
					resource.TestCheckResourceAttr("unifi_network.test", "domain_name", "test.local"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNetworkResourceConfig_dhcpNtp(name string, vlanID int, ntpServers []string) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name             = %q
  purpose          = "corporate"
  vlan_id          = %d
  subnet           = "10.%d.0.1/24"
  dhcp_enabled     = true
  dhcp_start       = "10.%d.0.10"
  dhcp_stop        = "10.%d.0.254"
  dhcp_ntp_enabled = true
  dhcp_ntp         = [%s]
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, formatStringListForHCL(ntpServers))
}

func testAccNetworkResourceConfig_dhcpGateway(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name                 = %q
  purpose              = "corporate"
  vlan_id              = %d
  subnet               = "10.%d.0.1/24"
  dhcp_enabled         = true
  dhcp_start           = "10.%d.0.10"
  dhcp_stop            = "10.%d.0.254"
  dhcp_gateway_enabled = true
  dhcp_gateway         = "10.%d.0.254"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_dhcpBoot(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name               = %q
  purpose            = "corporate"
  vlan_id            = %d
  subnet             = "10.%d.0.1/24"
  dhcp_enabled       = true
  dhcp_start         = "10.%d.0.10"
  dhcp_stop          = "10.%d.0.254"
  dhcp_boot_enabled  = true
  dhcp_boot_server   = "10.%d.0.10"
  dhcp_boot_filename = "pxelinux.0"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_dhcpAdvanced(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name                     = %q
  purpose                  = "corporate"
  vlan_id                  = %d
  subnet                   = "10.%d.0.1/24"
  dhcp_enabled             = true
  dhcp_start               = "10.%d.0.10"
  dhcp_stop                = "10.%d.0.254"
  dhcp_unifi_controller    = "10.%d.0.1"
  dhcp_wpad_url            = "http://wpad.test.local/wpad.dat"
  dhcp_time_offset_enabled = true
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_networkAccess(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name                         = %q
  purpose                      = "corporate"
  vlan_id                      = %d
  subnet                       = "10.%d.0.1/24"
  dhcp_enabled                 = true
  dhcp_start                   = "10.%d.0.10"
  dhcp_stop                    = "10.%d.0.254"
  internet_access_enabled      = false
  intra_network_access_enabled = false
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_mdns(name string, vlanID int) string {
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
  mdns_enabled = true
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_allDhcpOptions(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name                  = %q
  purpose               = "corporate"
  vlan_id               = %d
  subnet                = "10.%d.0.1/24"
  dhcp_enabled          = true
  dhcp_start            = "10.%d.0.10"
  dhcp_stop             = "10.%d.0.254"
  dhcp_lease            = 43200
  dhcp_dns              = ["8.8.8.8", "8.8.4.4"]
  dhcp_ntp_enabled      = true
  dhcp_ntp              = ["129.6.15.28"]
  dhcp_gateway_enabled  = true
  dhcp_gateway          = "10.%d.0.254"
  dhcp_boot_enabled     = true
  dhcp_boot_server      = "10.%d.0.10"
  dhcp_boot_filename    = "pxelinux.0"
  dhcp_unifi_controller = "10.%d.0.1"
  domain_name           = "test.local"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, vlanID%256, vlanID%256, vlanID%256)
}
