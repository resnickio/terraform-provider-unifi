package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckWanSlas(t *testing.T) {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		return
	}
	slas, err := client.ListWanSlas(context.Background())
	if err != nil || len(slas) == 0 {
		t.Skip("No WAN SLAs available on this controller")
	}
}

func TestAccWanSlaDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckWanSlas(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWanSlaDataSourceConfig_byName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_wan_sla.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_wan_sla.test", "name"),
				),
			},
		},
	})
}

func testAccWanSlaDataSourceConfig_byName() string {
	return testAccProviderConfig + `
data "unifi_wan_sla" "test" {
  name = "Default"
}
`
}
