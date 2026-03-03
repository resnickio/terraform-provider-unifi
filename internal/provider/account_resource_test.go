package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccountResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountResourceConfig("tf-acc-test-account", "testpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_account.test", "id"),
					resource.TestCheckResourceAttr("unifi_account.test", "name", "tf-acc-test-account"),
					resource.TestCheckResourceAttrSet("unifi_account.test", "site_id"),
				),
			},
			{
				ResourceName:            "unifi_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"x_password"},
			},
		},
	})
}

func TestAccAccountResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountResourceConfig("tf-acc-test-account-update", "testpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_account.test", "name", "tf-acc-test-account-update"),
				),
			},
			{
				Config: testAccAccountResourceConfig_withVLAN("tf-acc-test-account-update", "testpass456", 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_account.test", "vlan", "100"),
				),
			},
		},
	})
}

func testAccAccountResourceConfig(name, password string) string {
	return fmt.Sprintf(`
%s

resource "unifi_account" "test" {
  name       = %q
  x_password = %q
}
`, testAccProviderConfig, name, password)
}

func testAccAccountResourceConfig_withVLAN(name, password string, vlan int) string {
	return fmt.Sprintf(`
%s

resource "unifi_account" "test" {
  name       = %q
  x_password = %q
  vlan       = %d
}
`, testAccProviderConfig, name, password, vlan)
}
