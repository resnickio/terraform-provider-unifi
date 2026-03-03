package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckActiveClients(t *testing.T) string {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		t.Skip("No client available")
		return ""
	}
	clients, err := client.ListActiveClients(context.Background())
	if err != nil || len(clients) == 0 {
		t.Skip("No active clients available on this controller")
		return ""
	}
	return clients[0].MAC
}

func TestAccActiveClientDataSource_byMAC(t *testing.T) {
	mac := testAccPreCheckActiveClients(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActiveClientDataSourceConfig_byMAC(mac),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_active_client.test", "mac"),
				),
			},
		},
	})
}

func testAccActiveClientDataSourceConfig_byMAC(mac string) string {
	return testAccProviderConfig + fmt.Sprintf(`
data "unifi_active_client" "test" {
  mac = %q
}
`, mac)
}
