package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortProfileResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileResourceConfig_basic("tf-acc-test-port-profile-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-basic"),
					resource.TestCheckResourceAttrSet("unifi_port_profile.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_port_profile.test", "site_id"),
				),
			},
			{
				ResourceName:      "unifi_port_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortProfileResource_nativeVlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileResourceConfig_nativeVlan("tf-acc-test-port-profile-native", 3920),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-native"),
					resource.TestCheckResourceAttrSet("unifi_port_profile.test", "native_network_id"),
				),
			},
			{
				ResourceName:      "unifi_port_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortProfileResource_customTaggedVlans(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileResourceConfig_customTaggedVlans("tf-acc-test-port-profile-custom", 3921),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-custom"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "tagged_vlan_mgmt", "custom"),
					resource.TestCheckResourceAttrSet("unifi_port_profile.test", "native_network_id"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "excluded_network_ids.#", "1"),
				),
			},
			{
				ResourceName:      "unifi_port_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortProfileResource_poeAndIsolation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileResourceConfig_poeIsolation("tf-acc-test-port-profile-poe", 3926),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-poe"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "poe_mode", "auto"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "isolation", "true"),
				),
			},
			{
				ResourceName:      "unifi_port_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortProfileResource_stormControl(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileResourceConfig_stormControl("tf-acc-test-port-profile-storm"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-storm"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "stormctrl_bcast_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "stormctrl_bcast_rate", "1000"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "stormctrl_mcast_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "stormctrl_mcast_rate", "1000"),
				),
			},
			{
				ResourceName:      "unifi_port_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortProfileResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileResourceConfig_nativeVlan("tf-acc-test-port-profile-update", 3927),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-update"),
					resource.TestCheckResourceAttrSet("unifi_port_profile.test", "native_network_id"),
				),
			},
			{
				Config: testAccPortProfileResourceConfig_updateToCustom("tf-acc-test-port-profile-update-renamed", 3927),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", "tf-acc-test-port-profile-update-renamed"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "tagged_vlan_mgmt", "custom"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "excluded_network_ids.#", "1"),
				),
			},
		},
	})
}

func testAccPortProfileResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_profile" "test" {
  name = %q
}
`, testAccProviderConfig, name)
}

func testAccPortProfileResourceConfig_nativeVlan(name string, vlanBase int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "native" {
  name    = "%s-native-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_port_profile" "test" {
  name              = %q
  native_network_id = unifi_network.native.id
}
`, testAccProviderConfig, name, vlanBase, vlanBase%256, name)
}

func testAccPortProfileResourceConfig_customTaggedVlans(name string, vlanBase int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "native" {
  name    = "%s-native-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_network" "allowed" {
  name    = "%s-allowed-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_network" "excluded" {
  name    = "%s-excluded-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_port_profile" "test" {
  name                 = %q
  native_network_id    = unifi_network.native.id
  tagged_vlan_mgmt     = "custom"
  excluded_network_ids = [unifi_network.excluded.id]
}
`, testAccProviderConfig, name, vlanBase, vlanBase%256, name, vlanBase+1, (vlanBase+1)%256, name, vlanBase+2, (vlanBase+2)%256, name)
}

func testAccPortProfileResourceConfig_poeIsolation(name string, vlanBase int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "security" {
  name    = "%s-security-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_port_profile" "test" {
  name              = %q
  native_network_id = unifi_network.security.id
  poe_mode          = "auto"
  isolation         = true
}
`, testAccProviderConfig, name, vlanBase, vlanBase%256, name)
}

func testAccPortProfileResourceConfig_stormControl(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_profile" "test" {
  name                    = %q
  stormctrl_bcast_enabled = true
  stormctrl_bcast_rate    = 1000
  stormctrl_mcast_enabled = true
  stormctrl_mcast_rate    = 1000
}
`, testAccProviderConfig, name)
}

func testAccPortProfileResourceConfig_updateToCustom(name string, vlanBase int) string {
	return fmt.Sprintf(`
%s

resource "unifi_network" "native" {
  name    = "tf-acc-test-port-profile-update-native-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_network" "excluded" {
  name    = "tf-acc-test-port-profile-update-excluded-net"
  purpose = "corporate"
  vlan_id = %d
  subnet  = "10.%d.0.1/24"
}

resource "unifi_port_profile" "test" {
  name                 = %q
  native_network_id    = unifi_network.native.id
  tagged_vlan_mgmt     = "custom"
  excluded_network_ids = [unifi_network.excluded.id]
}
`, testAccProviderConfig, vlanBase, vlanBase%256, vlanBase+1, (vlanBase+1)%256, name)
}
