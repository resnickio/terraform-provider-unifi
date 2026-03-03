package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContentFilteringResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContentFilteringResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_content_filtering.test", "id", "content_filtering"),
					resource.TestCheckResourceAttr("unifi_content_filtering.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccContentFilteringResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContentFilteringResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_content_filtering.test", "enabled", "false"),
				),
			},
			{
				Config: testAccContentFilteringResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_content_filtering.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_content_filtering.test", "blocked_domains.#", "1"),
				),
			},
		},
	})
}

func testAccContentFilteringResourceConfig_basic() string {
	return testAccProviderConfig + `
resource "unifi_content_filtering" "test" {
  enabled = false
}
`
}

func testAccContentFilteringResourceConfig_updated() string {
	return testAccProviderConfig + `
resource "unifi_content_filtering" "test" {
  enabled        = true
  blocked_domains = ["example.com"]
}
`
}
