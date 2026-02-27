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

func TestAccNetworkResource_dhcpDnsEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_dhcpDnsEnabled("tf-acc-test-network-dns-en", 3916, true, []string{"8.8.8.8"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "1"),
					resource.TestCheckTypeSetElemAttr("unifi_network.test", "dhcp_dns.*", "8.8.8.8"),
				),
			},
			{
				Config: testAccNetworkResourceConfig_dhcpDnsEnabled("tf-acc-test-network-dns-en", 3916, false, []string{"8.8.8.8"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns_enabled", "false"),
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

func TestAccNetworkResource_dhcpTftpServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_dhcpTftpServer("tf-acc-test-network-tftp", 3917),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-tftp"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_server", "10.77.0.10"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_tftp_server", "10.77.0.11"),
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
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "2"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_ntp.#", "1"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_gateway_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_gateway", "10.75.0.254"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_boot_server", "10.75.0.10"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_tftp_server", "10.75.0.11"),
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

func testAccNetworkResourceConfig_dhcpDnsEnabled(name string, vlanID int, enabled bool, dnsServers []string) string {
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
  dhcp_dns_enabled = %t
  dhcp_dns         = [%s]
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, enabled, formatStringListForHCL(dnsServers))
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

func testAccNetworkResourceConfig_dhcpTftpServer(name string, vlanID int) string {
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
  dhcp_tftp_server   = "10.%d.0.11"
  dhcp_boot_filename = "pxelinux.0"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, vlanID%256, vlanID%256)
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
  dhcp_dns_enabled      = true
  dhcp_dns              = ["8.8.8.8", "8.8.4.4"]
  dhcp_ntp_enabled      = true
  dhcp_ntp              = ["129.6.15.28"]
  dhcp_gateway_enabled  = true
  dhcp_gateway          = "10.%d.0.254"
  dhcp_boot_enabled     = true
  dhcp_boot_server      = "10.%d.0.10"
  dhcp_tftp_server      = "10.%d.0.11"
  dhcp_boot_filename    = "pxelinux.0"
  dhcp_unifi_controller = "10.%d.0.1"
  domain_name           = "test.local"
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256, vlanID%256, vlanID%256, vlanID%256, vlanID%256)
}

// IPv6 tests

func TestAccNetworkResource_ipv6Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6Basic("tf-acc-test-network-ipv6", 3920),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-ipv6"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "auto"),
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

func TestAccNetworkResource_ipv6PrefixDelegation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6PD("tf-acc-test-network-ipv6pd", 3921),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-ipv6pd"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.interface_type", "pd"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_priority", "high"),
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

func TestAccNetworkResource_ipv6DHCPv6(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6DHCPv6("tf-acc-test-network-dhcpv6", 3922),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-dhcpv6"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.interface_type", "static"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.#", "2"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.0", "2001:4860:4860::8888"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.1", "2001:4860:4860::8844"),
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

func TestAccNetworkResource_ipv6Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6Basic("tf-acc-test-network-ipv6u", 3923),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "auto"),
				),
			},
			{
				Config: testAccNetworkResourceConfig_ipv6PD("tf-acc-test-network-ipv6u", 3923),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.interface_type", "pd"),
				),
			},
		},
	})
}

func testAccNetworkResourceConfig_ipv6Basic(name string, vlanID int) string {
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

  ipv6 = {
    setting_preference = "auto"
  }
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_ipv6PD(name string, vlanID int) string {
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

  ipv6 = {
    setting_preference = "manual"
    interface_type     = "pd"
    pd_interface       = "wan"
    pd_start           = "::2"
    pd_stop            = "::7d1"
    ra_enabled         = true
    ra_priority        = "high"
  }
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_ipv6DHCPv6(name string, vlanID int) string {
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

  ipv6 = {
    setting_preference  = "manual"
    interface_type      = "static"
    subnet              = "fd00:3922::1/64"
    wan_delegation_type = "none"
    dhcpv6_enabled      = true
    dhcpv6_start        = "fd00:3922::2"
    dhcpv6_stop         = "fd00:3922::ff"
    dhcpv6_dns          = ["2001:4860:4860::8888", "2001:4860:4860::8844"]
  }
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func TestAccNetworkResource_ipv6PDFull(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6PDFull("tf-acc-test-network-ipv6pdfull", 3924),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-ipv6pdfull"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.interface_type", "pd"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.pd_interface", "wan"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.pd_prefixid", "0"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.pd_start", "::2"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.pd_stop", "::7d1"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.pd_auto_prefixid_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_priority", "medium"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_preferred_lifetime", "14400"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_valid_lifetime", "86400"),
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

func TestAccNetworkResource_ipv6DHCPv6Full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6DHCPv6Full("tf-acc-test-network-dhcpv6full", 3925),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-dhcpv6full"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.interface_type", "static"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_start", "fd00:3925::2"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_stop", "fd00:3925::ff"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_lease_time", "43200"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns_auto", "false"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.#", "4"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.0", "2001:4860:4860::8888"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.1", "2001:4860:4860::8844"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.2", "2606:4700:4700::1111"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_dns.3", "2606:4700:4700::1001"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.dhcpv6_allow_slaac", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.ra_enabled", "true"),
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

func TestAccNetworkResource_ipv6StaticSubnet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkResourceConfig_ipv6Static("tf-acc-test-network-ipv6static", 3926),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "name", "tf-acc-test-network-ipv6static"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.setting_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.interface_type", "static"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.subnet", "fd00:abcd::1/64"),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6.wan_delegation_type", "none"),
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

func testAccNetworkResourceConfig_ipv6PDFull(name string, vlanID int) string {
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

  ipv6 = {
    setting_preference       = "manual"
    interface_type           = "pd"
    pd_interface             = "wan"
    pd_prefixid              = "0"
    pd_start                 = "::2"
    pd_stop                  = "::7d1"
    pd_auto_prefixid_enabled = false
    ra_enabled               = true
    ra_priority              = "medium"
    ra_preferred_lifetime    = 14400
    ra_valid_lifetime        = 86400
  }
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_ipv6DHCPv6Full(name string, vlanID int) string {
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

  ipv6 = {
    setting_preference  = "manual"
    interface_type      = "static"
    subnet              = "fd00:3925::1/64"
    wan_delegation_type = "none"
    dhcpv6_enabled      = true
    dhcpv6_start        = "fd00:3925::2"
    dhcpv6_stop         = "fd00:3925::ff"
    dhcpv6_lease_time   = 43200
    dhcpv6_dns_auto     = false
    dhcpv6_dns          = ["2001:4860:4860::8888", "2001:4860:4860::8844", "2606:4700:4700::1111", "2606:4700:4700::1001"]
    dhcpv6_allow_slaac  = true
    ra_enabled          = true
  }
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkResourceConfig_ipv6Static(name string, vlanID int) string {
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

  ipv6 = {
    setting_preference  = "manual"
    interface_type      = "static"
    subnet              = "fd00:abcd::1/64"
    wan_delegation_type = "none"
  }
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}
