package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSiteDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteDataSourceConfig_byName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_site.test", "id"),
					resource.TestCheckResourceAttr("data.unifi_site.test", "name", "default"),
					resource.TestCheckResourceAttrSet("data.unifi_site.test", "description"),
				),
			},
		},
	})
}

func testAccSiteDataSourceConfig_byName() string {
	return testAccProviderConfig + `
data "unifi_site" "test" {
  name = "default"
}
`
}
