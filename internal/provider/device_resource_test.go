package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceResource_basic(t *testing.T) {
	mac := testAccGetFirstDeviceMAC(t)
	if mac == "" {
		t.Skip("No device available for testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckDevice(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceResourceConfig_basic(mac),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_device.test", "id"),
					resource.TestCheckResourceAttr("unifi_device.test", "mac", mac),
					resource.TestCheckResourceAttrSet("unifi_device.test", "type"),
					resource.TestCheckResourceAttrSet("unifi_device.test", "model"),
				),
			},
			{
				ResourceName:            "unifi_device.test",
				ImportState:             true,
				ImportStateId:           mac,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func TestAccDeviceResource_updateName(t *testing.T) {
	mac := testAccGetFirstDeviceMAC(t)
	if mac == "" {
		t.Skip("No device available for testing")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckDevice(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceResourceConfig_withName(mac, "tf-acc-test-device"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_device.test", "name", "tf-acc-test-device"),
				),
			},
			{
				Config: testAccDeviceResourceConfig_withName(mac, "tf-acc-test-device-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_device.test", "name", "tf-acc-test-device-updated"),
				),
			},
		},
	})
}

func testAccDeviceResourceConfig_basic(mac string) string {
	return fmt.Sprintf(`
%s

resource "unifi_device" "test" {
  mac = %q
}
`, testAccProviderConfig, mac)
}

func testAccDeviceResourceConfig_withName(mac, name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_device" "test" {
  mac  = %q
  name = %q
}
`, testAccProviderConfig, mac, name)
}
