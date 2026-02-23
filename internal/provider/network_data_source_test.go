package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDataSourceConfig_byName("tf-acc-test-ds-network", 3950),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_network.test", "name", "tf-acc-test-ds-network"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "purpose", "corporate"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "vlan_id", "3950"),
					resource.TestCheckResourceAttrSet("data.unifi_network.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_network.test", "site_id"),
				),
			},
		},
	})
}

func TestAccNetworkDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDataSourceConfig_byID("tf-acc-test-ds-network-id", 3951),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_network.test", "name", "tf-acc-test-ds-network-id"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "purpose", "corporate"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "vlan_id", "3951"),
					resource.TestCheckResourceAttrSet("data.unifi_network.test", "id"),
				),
			},
		},
	})
}

func TestAccNetworkDataSource_allAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkDataSourceConfig_full("tf-acc-test-ds-network-full", 3952),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_network.test", "name", "tf-acc-test-ds-network-full"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "purpose", "corporate"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "vlan_id", "3952"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "subnet", "10.39.52.1/24"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_enabled", "true"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_start", "10.39.52.10"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_stop", "10.39.52.254"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_dns_enabled", "true"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_dns.#", "2"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_boot_enabled", "true"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_boot_server", "10.39.52.10"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_tftp_server", "10.39.52.11"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcp_boot_filename", "pxelinux.0"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "domain_name", "dstest.local"),
				),
			},
		},
	})
}

func testAccNetworkDataSourceConfig_byName(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "source" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
}

data "unifi_network" "test" {
  name = unifi_network.source.name
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkDataSourceConfig_byID(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "source" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
}

data "unifi_network" "test" {
  id = unifi_network.source.id
}
`, testAccProviderConfig, name, vlanID, vlanID%256, vlanID%256, vlanID%256)
}

func testAccNetworkDataSourceConfig_full(name string, vlanID int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "source" {
  name         = %q
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.39.52.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.39.52.10"
  dhcp_stop    = "10.39.52.254"
  dhcp_dns_enabled = true
  dhcp_dns         = ["8.8.8.8", "8.8.4.4"]
  dhcp_boot_enabled  = true
  dhcp_boot_server   = "10.39.52.10"
  dhcp_tftp_server   = "10.39.52.11"
  dhcp_boot_filename = "pxelinux.0"
  domain_name  = "dstest.local"
}

data "unifi_network" "test" {
  name = unifi_network.source.name
}
`, testAccProviderConfig, name, vlanID)
}
