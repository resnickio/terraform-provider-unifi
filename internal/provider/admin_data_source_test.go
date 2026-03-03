package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAdminDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAdminDataSourceConfig_byName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_admin.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_admin.test", "email"),
				),
			},
		},
	})
}

func testAccAdminDataSourceConfig_byName() string {
	return testAccProviderConfig + `
data "unifi_admin" "test" {
  name = "admin"
}
`
}
