package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevicePortOverrideResource_basic(t *testing.T) {
	testAccPreCheckSwitch(t)
	mac := testAccGetFirstSwitchMAC(t)
	if mac == "" {
		t.Skip("No switch MAC available")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePortOverrideResourceConfig_basic(mac, "tf-acc-test-port"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_device_port_override.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_device_port_override.test", "device_id"),
					resource.TestCheckResourceAttr("unifi_device_port_override.test", "port_idx", "1"),
					resource.TestCheckResourceAttr("unifi_device_port_override.test", "name", "tf-acc-test-port"),
				),
			},
			{
				ResourceName:      "unifi_device_port_override.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDevicePortOverrideResource_withProfile(t *testing.T) {
	testAccPreCheckSwitch(t)
	mac := testAccGetFirstSwitchMAC(t)
	if mac == "" {
		t.Skip("No switch MAC available")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePortOverrideResourceConfig_withProfile(mac, "tf-acc-test-port-profile"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_device_port_override.test", "id"),
					resource.TestCheckResourceAttr("unifi_device_port_override.test", "port_idx", "2"),
					resource.TestCheckResourceAttr("unifi_device_port_override.test", "name", "tf-acc-test-port-profile"),
					resource.TestCheckResourceAttrSet("unifi_device_port_override.test", "port_profile_id"),
				),
			},
		},
	})
}

func TestAccDevicePortOverrideResource_update(t *testing.T) {
	testAccPreCheckSwitch(t)
	mac := testAccGetFirstSwitchMAC(t)
	if mac == "" {
		t.Skip("No switch MAC available")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDevicePortOverrideResourceConfig_basic(mac, "tf-acc-test-port-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_device_port_override.test", "name", "tf-acc-test-port-update"),
				),
			},
			{
				Config: testAccDevicePortOverrideResourceConfig_basic(mac, "tf-acc-test-port-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_device_port_override.test", "name", "tf-acc-test-port-updated"),
				),
			},
		},
	})
}

func testAccDevicePortOverrideResourceConfig_basic(mac, name string) string {
	return fmt.Sprintf(`
%s

data "unifi_device" "test" {
  mac = %q
}

resource "unifi_device_port_override" "test" {
  device_id = data.unifi_device.test.id
  port_idx  = 1
  name      = %q
}
`, testAccProviderConfig, mac, name)
}

func testAccDevicePortOverrideResourceConfig_withProfile(mac, name string) string {
	return fmt.Sprintf(`
%s

data "unifi_device" "test" {
  mac = %q
}

resource "unifi_port_profile" "test" {
  name = "tf-acc-test-port-profile-override"
}

resource "unifi_device_port_override" "test" {
  device_id       = data.unifi_device.test.id
  port_idx        = 2
  name            = %q
  port_profile_id = unifi_port_profile.test.id
}
`, testAccProviderConfig, mac, name)
}
