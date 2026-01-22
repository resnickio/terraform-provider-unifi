package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceDataSource_byMAC(t *testing.T) {
	testAccPreCheckDevice(t)
	mac := testAccGetFirstDeviceMAC(t)
	if mac == "" {
		t.Skip("No device MAC available")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceDataSourceConfig_byMAC(mac),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_device.test", "id"),
					resource.TestCheckResourceAttr("data.unifi_device.test", "mac", mac),
					resource.TestCheckResourceAttrSet("data.unifi_device.test", "type"),
				),
			},
		},
	})
}

func testAccDeviceDataSourceConfig_byMAC(mac string) string {
	return fmt.Sprintf(`
%s

data "unifi_device" "test" {
  mac = %q
}
`, testAccProviderConfig, mac)
}
