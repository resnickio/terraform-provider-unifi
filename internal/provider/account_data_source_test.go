package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccountDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_byName("tf-acc-test-account-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_account.test", "name", "tf-acc-test-account-ds"),
					resource.TestCheckResourceAttrSet("data.unifi_account.test", "id"),
				),
			},
		},
	})
}

func testAccAccountDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_account" "test" {
  name       = %q
  x_password = "testpass123"
}

data "unifi_account" "test" {
  name = unifi_account.test.name
}
`, testAccProviderConfig, name)
}
