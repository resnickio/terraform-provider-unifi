package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSiteResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteResourceConfig("tf-acc-test-site"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_site.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_site.test", "name"),
					resource.TestCheckResourceAttr("unifi_site.test", "description", "tf-acc-test-site"),
				),
			},
			{
				ResourceName:      "unifi_site.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSiteResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteResourceConfig("tf-acc-test-site-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_site.test", "description", "tf-acc-test-site-update"),
				),
			},
			{
				Config: testAccSiteResourceConfig("tf-acc-test-site-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_site.test", "description", "tf-acc-test-site-updated"),
				),
			},
		},
	})
}

func testAccSiteResourceConfig(desc string) string {
	return fmt.Sprintf(`
%s

resource "unifi_site" "test" {
  description = %q
}
`, testAccProviderConfig, desc)
}
