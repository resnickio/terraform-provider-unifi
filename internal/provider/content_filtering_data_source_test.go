package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckContentFiltering(t *testing.T) {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		return
	}
	_, err := client.GetContentFiltering(context.Background())
	if err != nil {
		t.Skipf("Content filtering not available on this controller: %v", err)
	}
}

func TestAccContentFilteringDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckContentFiltering(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContentFilteringDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_content_filtering.test", "enabled"),
				),
			},
		},
	})
}

func testAccContentFilteringDataSourceConfig() string {
	return testAccProviderConfig + `
data "unifi_content_filtering" "test" {}
`
}
