package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallGroupResource_addressGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-addr-group", []string{"10.0.0.1", "10.0.0.2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-addr-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "group_type", "address-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "2"),
					resource.TestCheckResourceAttrSet("unifi_firewall_group.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_firewall_group.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_portGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccFirewallGroupResourceConfig_portGroup("tf-acc-test-port-group", []string{"80", "443", "8080-8090"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-port-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "group_type", "port-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "3"),
					resource.TestCheckResourceAttrSet("unifi_firewall_group.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-update-group", []string{"10.0.0.1"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-update-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "1"),
				),
			},
			// Update - add members and change name
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-update-group-renamed", []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-update-group-renamed"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "3"),
				),
			},
		},
	})
}

func TestAccFirewallGroupResource_singleMember(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with single member
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-single", []string{"192.168.1.1"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-single"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "1"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.0", "192.168.1.1"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_cidrMembers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with CIDR members
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-cidr", []string{"10.0.0.0/24", "192.168.1.0/24", "172.16.0.0/16"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-cidr"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "group_type", "address-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "3"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_mixedAddresses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with mixed IPs and CIDRs
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-mixed", []string{"10.0.0.1", "192.168.1.0/24", "172.16.0.100"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-mixed"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "3"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_portRanges(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with port ranges
			{
				Config: testAccFirewallGroupResourceConfig_portGroup("tf-acc-test-port-ranges", []string{"1000-2000", "3000-4000", "5000-6000"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-port-ranges"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "group_type", "port-group"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "3"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_mixedPorts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with mixed single ports and ranges
			{
				Config: testAccFirewallGroupResourceConfig_portGroup("tf-acc-test-mixed-ports", []string{"22", "80", "443", "8000-9000"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-mixed-ports"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "4"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_updateMembers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-update-members", []string{"10.0.0.1", "10.0.0.2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "2"),
				),
			},
			// Update - remove a member
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-update-members", []string{"10.0.0.1"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "1"),
				),
			},
			// Update - add more members
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-update-members", []string{"10.0.0.1", "10.0.0.5", "10.0.0.10", "10.0.0.15"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "4"),
				),
			},
			// Update - replace all members
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-update-members", []string{"192.168.0.1", "192.168.0.2"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "2"),
				),
			},
		},
	})
}

func TestAccFirewallGroupResource_largeAddressGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with many members
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-large", []string{
					"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5",
					"10.0.0.6", "10.0.0.7", "10.0.0.8", "10.0.0.9", "10.0.0.10",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-large"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "10"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_largePortGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with many port entries
			{
				Config: testAccFirewallGroupResourceConfig_portGroup("tf-acc-test-large-ports", []string{
					"22", "80", "443", "3306", "5432", "6379", "8080", "8443", "9000", "9090",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-large-ports"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "10"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_commonPorts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with commonly used ports
			{
				Config: testAccFirewallGroupResourceConfig_portGroup("tf-acc-test-common-ports", []string{
					"22",    // SSH
					"80",    // HTTP
					"443",   // HTTPS
					"25",    // SMTP
					"53",    // DNS
					"3389",  // RDP
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-common-ports"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "6"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallGroupResource_privateNetworks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with RFC1918 private networks
			{
				Config: testAccFirewallGroupResourceConfig_addressGroup("tf-acc-test-rfc1918", []string{
					"10.0.0.0/8",
					"172.16.0.0/12",
					"192.168.0.0/16",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "name", "tf-acc-test-rfc1918"),
					resource.TestCheckResourceAttr("unifi_firewall_group.test", "members.#", "3"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFirewallGroupResourceConfig_addressGroup(name string, members []string) string {
	membersStr := ""
	for i, m := range members {
		if i > 0 {
			membersStr += ", "
		}
		membersStr += fmt.Sprintf("%q", m)
	}

	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "test" {
  name       = %q
  group_type = "address-group"
  members    = [%s]
}
`, testAccProviderConfig, name, membersStr)
}

func testAccFirewallGroupResourceConfig_portGroup(name string, members []string) string {
	membersStr := ""
	for i, m := range members {
		if i > 0 {
			membersStr += ", "
		}
		membersStr += fmt.Sprintf("%q", m)
	}

	return fmt.Sprintf(`
%s

resource "unifi_firewall_group" "test" {
  name       = %q
  group_type = "port-group"
  members    = [%s]
}
`, testAccProviderConfig, name, membersStr)
}
