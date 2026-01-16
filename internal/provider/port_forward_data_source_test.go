package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortForwardDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortForwardDataSourceConfig_byName("tf-acc-test-pf-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_port_forward.test", "name", "tf-acc-test-pf-ds"),
					resource.TestCheckResourceAttr("data.unifi_port_forward.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("data.unifi_port_forward.test", "dst_port", "8080"),
					resource.TestCheckResourceAttr("data.unifi_port_forward.test", "fwd_ip", "192.168.1.100"),
					resource.TestCheckResourceAttrSet("data.unifi_port_forward.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_port_forward.test", "site_id"),
				),
			},
		},
	})
}

func TestAccPortForwardDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortForwardDataSourceConfig_byID("tf-acc-test-pf-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_port_forward.test", "name", "tf-acc-test-pf-ds-id"),
					resource.TestCheckResourceAttrSet("data.unifi_port_forward.test", "id"),
				),
			},
		},
	})
}

func testAccPortForwardDataSourceConfig_byName(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = "8080"
  fwd_port = "80"
  fwd_ip   = "192.168.1.100"
}

data "unifi_port_forward" "test" {
  name = unifi_port_forward.test.name
}
`, testAccProviderConfig, name)
}

func testAccPortForwardDataSourceConfig_byID(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = "8081"
  fwd_port = "81"
  fwd_ip   = "192.168.1.101"
}

data "unifi_port_forward" "test" {
  id = unifi_port_forward.test.id
}
`, testAccProviderConfig, name)
}
