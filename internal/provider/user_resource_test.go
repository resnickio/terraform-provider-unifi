package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("aa:bb:cc:00:00:01", "tf-acc-test-user-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:01"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-basic"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "site_id"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_fixedIP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_fixedIP(
					"aa:bb:cc:00:00:02",
					"tf-acc-test-user-fixedip",
					"192.168.3.100",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:02"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-fixedip"),
					resource.TestCheckResourceAttr("unifi_user.test", "use_fixed_ip", "true"),
					resource.TestCheckResourceAttr("unifi_user.test", "fixed_ip", "192.168.3.100"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "network_id"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_localDNS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_localDNS(
					"aa:bb:cc:00:00:03",
					"tf-acc-test-user-dns",
					"tf-acc-test-host",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:03"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-dns"),
					resource.TestCheckResourceAttr("unifi_user.test", "local_dns_record", "tf-acc-test-host"),
					resource.TestCheckResourceAttr("unifi_user.test", "local_dns_record_enabled", "true"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_blocked(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_blocked("aa:bb:cc:00:00:04", "tf-acc-test-user-blocked"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:04"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-blocked"),
					resource.TestCheckResourceAttr("unifi_user.test", "blocked", "true"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_userGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_userGroup(
					"aa:bb:cc:00:00:05",
					"tf-acc-test-user-group-assign",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:05"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-group-assign"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "usergroup_id"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_full(
					"aa:bb:cc:00:00:06",
					"tf-acc-test-user-full",
					"192.168.3.106",
					"tf-acc-test-fullhost",
					"Full test user note",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:06"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-full"),
					resource.TestCheckResourceAttr("unifi_user.test", "use_fixed_ip", "true"),
					resource.TestCheckResourceAttr("unifi_user.test", "fixed_ip", "192.168.3.106"),
					resource.TestCheckResourceAttr("unifi_user.test", "local_dns_record", "tf-acc-test-fullhost"),
					resource.TestCheckResourceAttr("unifi_user.test", "local_dns_record_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_user.test", "note", "Full test user note"),
					resource.TestCheckResourceAttr("unifi_user.test", "noted", "true"),
					resource.TestCheckResourceAttr("unifi_user.test", "blocked", "false"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "network_id"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "usergroup_id"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_user.test", "site_id"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_basic("aa:bb:cc:00:00:07", "tf-acc-test-user-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-update"),
				),
			},
			{
				Config: testAccUserResourceConfig_fixedIP(
					"aa:bb:cc:00:00:07",
					"tf-acc-test-user-updated",
					"192.168.3.107",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-updated"),
					resource.TestCheckResourceAttr("unifi_user.test", "use_fixed_ip", "true"),
					resource.TestCheckResourceAttr("unifi_user.test", "fixed_ip", "192.168.3.107"),
				),
			},
		},
	})
}

func TestAccUserResource_note(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig_note("aa:bb:cc:00:00:08", "tf-acc-test-user-note", "Test note content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user.test", "mac", "aa:bb:cc:00:00:08"),
					resource.TestCheckResourceAttr("unifi_user.test", "name", "tf-acc-test-user-note"),
					resource.TestCheckResourceAttr("unifi_user.test", "note", "Test note content"),
					resource.TestCheckResourceAttr("unifi_user.test", "noted", "true"),
				),
			},
			{
				ResourceName:      "unifi_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserResourceConfig_basic(mac, name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user" "test" {
  mac  = %q
  name = %q
}
`, testAccProviderConfig, mac, name)
}

func testAccUserResourceConfig_fixedIP(mac, name, fixedIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name    = "tf-acc-test-user-net"
  purpose = "corporate"
  subnet  = "192.168.3.0/24"
  vlan_id = 3900
}

resource "unifi_user" "test" {
  mac          = %q
  name         = %q
  use_fixed_ip = true
  fixed_ip     = %q
  network_id   = unifi_network.test.id
}
`, testAccProviderConfig, mac, name, fixedIP)
}

func testAccUserResourceConfig_localDNS(mac, name, dnsRecord string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user" "test" {
  mac                     = %q
  name                    = %q
  local_dns_record        = %q
  local_dns_record_enabled = true
}
`, testAccProviderConfig, mac, name, dnsRecord)
}

func testAccUserResourceConfig_blocked(mac, name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user" "test" {
  mac     = %q
  name    = %q
  blocked = true
}
`, testAccProviderConfig, mac, name)
}

func testAccUserResourceConfig_userGroup(mac, name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name              = "tf-acc-test-user-ug"
  qos_rate_max_down = 10000
  qos_rate_max_up   = 5000
}

resource "unifi_user" "test" {
  mac          = %q
  name         = %q
  usergroup_id = unifi_user_group.test.id
}
`, testAccProviderConfig, mac, name)
}

func testAccUserResourceConfig_full(mac, name, fixedIP, dnsRecord, note string) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "test" {
  name    = "tf-acc-test-user-net-full"
  purpose = "corporate"
  subnet  = "192.168.3.0/24"
  vlan_id = 3901
}

resource "unifi_user_group" "test" {
  name              = "tf-acc-test-user-ug-full"
  qos_rate_max_down = 50000
  qos_rate_max_up   = 25000
}

resource "unifi_user" "test" {
  mac                      = %q
  name                     = %q
  use_fixed_ip             = true
  fixed_ip                 = %q
  network_id               = unifi_network.test.id
  local_dns_record         = %q
  local_dns_record_enabled = true
  usergroup_id             = unifi_user_group.test.id
  note                     = %q
  noted                    = true
  blocked                  = false
}
`, testAccProviderConfig, mac, name, fixedIP, dnsRecord, note)
}

func testAccUserResourceConfig_note(mac, name, note string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user" "test" {
  mac   = %q
  name  = %q
  note  = %q
  noted = true
}
`, testAccProviderConfig, mac, name, note)
}
