package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckVpnConnections(t *testing.T) {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		return
	}
	connections, err := client.ListVpnConnections(context.Background())
	if err != nil || len(connections) == 0 {
		t.Skip("No VPN connections available on this controller")
	}
}

func TestAccVpnConnectionDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckVpnConnections(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnConnectionDataSourceConfig_byName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_vpn_connection.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_vpn_connection.test", "name"),
				),
			},
		},
	})
}

func testAccVpnConnectionDataSourceConfig_byName() string {
	return testAccProviderConfig + `
data "unifi_vpn_connection" "test" {
  name = "Default"
}
`
}
