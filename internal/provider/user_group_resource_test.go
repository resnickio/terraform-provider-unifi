package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroupResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccUserGroupResourceConfig_basic("tf-acc-test-user-group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group"),
					resource.TestCheckResourceAttrSet("unifi_user_group.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_user_group.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserGroupResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all options
			{
				Config: testAccUserGroupResourceConfig_full("tf-acc-test-user-group-full", 10000, 5000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group-full"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_down", "10000"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_up", "5000"),
					resource.TestCheckResourceAttrSet("unifi_user_group.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserGroupResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccUserGroupResourceConfig_full("tf-acc-test-user-group-update", 10000, 5000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group-update"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_down", "10000"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_up", "5000"),
				),
			},
			// Update - change limits
			{
				Config: testAccUserGroupResourceConfig_full("tf-acc-test-user-group-updated", 50000, 25000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group-updated"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_down", "50000"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_up", "25000"),
				),
			},
		},
	})
}

func TestAccUserGroupResource_unlimited(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with unlimited bandwidth (-1)
			{
				Config: testAccUserGroupResourceConfig_full("tf-acc-test-user-group-unlimited", -1, -1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group-unlimited"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_down", "-1"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_up", "-1"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserGroupResource_downloadOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with download limit only
			{
				Config: testAccUserGroupResourceConfig_downloadOnly("tf-acc-test-user-group-dl", 100000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group-dl"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_down", "100000"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserGroupResource_uploadOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with upload limit only
			{
				Config: testAccUserGroupResourceConfig_uploadOnly("tf-acc-test-user-group-ul", 50000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_user_group.test", "name", "tf-acc-test-user-group-ul"),
					resource.TestCheckResourceAttr("unifi_user_group.test", "qos_rate_max_up", "50000"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserGroupResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name = %q
}
`, testAccProviderConfig, name)
}

func testAccUserGroupResourceConfig_full(name string, maxDown, maxUp int) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name             = %q
  qos_rate_max_down = %d
  qos_rate_max_up   = %d
}
`, testAccProviderConfig, name, maxDown, maxUp)
}

func testAccUserGroupResourceConfig_downloadOnly(name string, maxDown int) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name             = %q
  qos_rate_max_down = %d
}
`, testAccProviderConfig, name, maxDown)
}

func testAccUserGroupResourceConfig_uploadOnly(name string, maxUp int) string {
	return fmt.Sprintf(`
%s

resource "unifi_user_group" "test" {
  name           = %q
  qos_rate_max_up = %d
}
`, testAccProviderConfig, name, maxUp)
}
