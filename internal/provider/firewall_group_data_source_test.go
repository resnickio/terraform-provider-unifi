package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallGroupDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupDataSourceConfig_byName("tf-acc-test-ds-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "name", "tf-acc-test-ds-group"),
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "group_type", "address-group"),
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "members.#", "2"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_group.test", "site_id"),
				),
			},
		},
	})
}

func TestAccFirewallGroupDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupDataSourceConfig_byID("tf-acc-test-ds-group-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "name", "tf-acc-test-ds-group-id"),
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "group_type", "address-group"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_group.test", "id"),
				),
			},
		},
	})
}

func TestAccFirewallGroupDataSource_portGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupDataSourceConfig_portGroup("tf-acc-test-ds-ports"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "name", "tf-acc-test-ds-ports"),
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "group_type", "port-group"),
					resource.TestCheckResourceAttr("data.unifi_firewall_group.test", "members.#", "3"),
					resource.TestCheckResourceAttrSet("data.unifi_firewall_group.test", "id"),
				),
			},
		},
	})
}

func testAccFirewallGroupDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "source" {
  name       = %q
  group_type = "address-group"
  members    = ["192.168.1.0/24", "10.0.0.0/8"]
}

data "unifi_firewall_group" "test" {
  name = unifi_firewall_group.source.name
}
`, testAccProviderConfig, name)
}

func testAccFirewallGroupDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "source" {
  name       = %q
  group_type = "address-group"
  members    = ["192.168.1.1"]
}

data "unifi_firewall_group" "test" {
  id = unifi_firewall_group.source.id
}
`, testAccProviderConfig, name)
}

func testAccFirewallGroupDataSourceConfig_portGroup(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "source" {
  name       = %q
  group_type = "port-group"
  members    = ["80", "443", "8080-8090"]
}

data "unifi_firewall_group" "test" {
  name = unifi_firewall_group.source.name
}
`, testAccProviderConfig, name)
}
