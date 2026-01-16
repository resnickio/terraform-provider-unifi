package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroupDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_byName("tf-acc-test-user-group-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_user_group.test", "name", "tf-acc-test-user-group-ds"),
					resource.TestCheckResourceAttrSet("data.unifi_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_user_group.test", "site_id"),
				),
			},
		},
	})
}

func TestAccUserGroupDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_byID("tf-acc-test-user-group-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_user_group.test", "name", "tf-acc-test-user-group-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_user_group.test", "id"),
				),
			},
		},
	})
}

func testAccUserGroupDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name = %q
}

data "unifi_user_group" "test" {
  name = unifi_user_group.test.name
}
`, testAccProviderConfig, name)
}

func testAccUserGroupDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name = %q
}

data "unifi_user_group" "test" {
  id = unifi_user_group.test.id
}
`, testAccProviderConfig, name)
}
