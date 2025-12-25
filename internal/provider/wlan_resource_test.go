package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func getDefaultAPGroupID(t *testing.T) string {
	client := testAccGetClient(t)
	if client == nil {
		return ""
	}

	wlans, err := client.ListWLANs(context.Background())
	if err != nil {
		t.Skipf("Could not list WLANs to find AP group: %v", err)
		return ""
	}

	for _, wlan := range wlans {
		if len(wlan.APGroupIDs) > 0 {
			return wlan.APGroupIDs[0]
		}
	}

	t.Skip("No AP groups found on controller - WLAN tests require at least one AP group")
	return ""
}

func TestAccWLANResource_basic(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_basic("tf-acc-test-wlan", "TestPassword123!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "security", "wpapsk"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_wlan.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_wlan.test", "site_id"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func TestAccWLANResource_full(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_full("tf-acc-test-wlan-full", "TestPassword123!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-full"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "security", "wpapsk"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "wpa_mode", "wpa2"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "hide_ssid", "false"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "is_guest", "false"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "l2_isolation", "true"),
					resource.TestCheckResourceAttrSet("unifi_wlan.test", "id"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func TestAccWLANResource_update(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_basic("tf-acc-test-wlan-update", "TestPassword123!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-update"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "hide_ssid", "false"),
				),
			},
			{
				Config: testAccWLANResourceConfig_hidden("tf-acc-test-wlan-updated", "TestPassword456!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-updated"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "hide_ssid", "true"),
				),
			},
		},
	})
}

func TestAccWLANResource_open(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_open("tf-acc-test-wlan-open", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-open"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "security", "open"),
				),
			},
			{
				ResourceName:      "unifi_wlan.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWLANResource_guest(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_guest("tf-acc-test-wlan-guest", "GuestPassword123!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-guest"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "is_guest", "true"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func TestAccWLANResource_disabled(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_disabled("tf-acc-test-wlan-disabled", "TestPassword123!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-disabled"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "enabled", "false"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func TestAccWLANResource_networkVlan(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_networkVlan("tf-acc-test-wlan-vlan", "TestPassword123!", 3960, apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-vlan"),
					resource.TestCheckResourceAttrSet("unifi_wlan.test", "network_id"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func TestAccWLANResource_wpa3(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_wpa3("tf-acc-test-wlan-wpa3", "TestPassword123!", apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-wpa3"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "wpa3_support", "true"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func TestAccWLANResource_macFilter(t *testing.T) {
	apGroupID := getDefaultAPGroupID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWLANResourceConfig_macFilter("tf-acc-test-wlan-mac", "TestPassword123!", []string{"00:11:22:33:44:55", "66:77:88:99:AA:BB"}, apGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.test", "name", "tf-acc-test-wlan-mac"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "mac_filter_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "mac_filter_policy", "allow"),
					resource.TestCheckResourceAttr("unifi_wlan.test", "mac_filter_list.#", "2"),
				),
			},
			{
				ResourceName:            "unifi_wlan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"passphrase"},
			},
		},
	})
}

func testAccWLANResourceConfig_basic(name, passphrase, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = %q
  ap_group_ids = [%q]
}
`, testAccProviderConfig, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_full(name, passphrase, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name                 = %q
  security             = "wpapsk"
  passphrase           = %q
  ap_group_ids         = [%q]
  wpa_mode             = "wpa2"
  enabled              = true
  hide_ssid            = false
  is_guest             = false
  l2_isolation         = true
  fast_roaming_enabled = false
  bss_transition       = true
  uapsd_enabled        = true
  pmf_mode             = "optional"
}
`, testAccProviderConfig, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_hidden(name, passphrase, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = %q
  ap_group_ids = [%q]
  hide_ssid    = true
}
`, testAccProviderConfig, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_open(name, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name         = %q
  security     = "open"
  ap_group_ids = [%q]
}
`, testAccProviderConfig, name, apGroupID)
}

func testAccWLANResourceConfig_guest(name, passphrase, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = %q
  ap_group_ids = [%q]
  is_guest     = true
}
`, testAccProviderConfig, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_disabled(name, passphrase, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = %q
  ap_group_ids = [%q]
  enabled      = false
}
`, testAccProviderConfig, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_networkVlan(name, passphrase string, vlan int, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "vlan_network" {
  name         = "%s-network"
  purpose      = "corporate"
  vlan_id      = %d
  subnet       = "10.%d.0.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.%d.0.10"
  dhcp_stop    = "10.%d.0.254"
}

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = %q
  ap_group_ids = [%q]
  network_id   = unifi_network.vlan_network.id
}
`, testAccProviderConfig, name, vlan, vlan%256, vlan%256, vlan%256, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_wpa3(name, passphrase, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name         = %q
  security     = "wpapsk"
  passphrase   = %q
  ap_group_ids = [%q]
  wpa_mode     = "wpa2"
  wpa3_support = true
  pmf_mode     = "required"
}
`, testAccProviderConfig, name, passphrase, apGroupID)
}

func testAccWLANResourceConfig_macFilter(name, passphrase string, macs []string, apGroupID string) string {
	return fmt.Sprintf(`
%s

resource "unifi_wlan" "test" {
  name               = %q
  security           = "wpapsk"
  passphrase         = %q
  ap_group_ids       = [%q]
  mac_filter_enabled = true
  mac_filter_policy  = "allow"
  mac_filter_list    = [%s]
}
`, testAccProviderConfig, name, passphrase, apGroupID, formatStringListForHCL(macs))
}
