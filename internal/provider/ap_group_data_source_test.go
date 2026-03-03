package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckAPGroups(t *testing.T) string {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		t.Skip("No client available")
		return ""
	}
	groups, err := client.ListAPGroups(context.Background())
	if err != nil || len(groups) == 0 {
		t.Skip("No AP groups available on this controller")
		return ""
	}
	return groups[0].Name
}

func TestAccAPGroupDataSource_byName(t *testing.T) {
	name := testAccPreCheckAPGroups(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPGroupDataSourceConfig_byName(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_ap_group.test", "id"),
					resource.TestCheckResourceAttr("data.unifi_ap_group.test", "name", name),
				),
			},
		},
	})
}

func TestAccAPGroupDataSource_byID(t *testing.T) {
	name := testAccPreCheckAPGroups(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPGroupDataSourceConfig_byID(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_ap_group.by_id", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_ap_group.by_id", "name"),
				),
			},
		},
	})
}

func testAccAPGroupDataSourceConfig_byName(name string) string {
	return testAccProviderConfig + fmt.Sprintf(`
data "unifi_ap_group" "test" {
  name = %q
}
`, name)
}

func testAccAPGroupDataSourceConfig_byID(name string) string {
	return testAccProviderConfig + fmt.Sprintf(`
data "unifi_ap_group" "default" {
  name = %q
}

data "unifi_ap_group" "by_id" {
  id = data.unifi_ap_group.default.id
}
`, name)
}
