package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStaticDNSDataSource_byKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSDataSourceConfig_byKey("tf-acc-test-static-dns-ds.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_static_dns.test", "key", "tf-acc-test-static-dns-ds.example.com"),
					resource.TestCheckResourceAttr("data.unifi_static_dns.test", "value", "192.168.1.100"),
					resource.TestCheckResourceAttr("data.unifi_static_dns.test", "record_type", "A"),
					resource.TestCheckResourceAttrSet("data.unifi_static_dns.test", "id"),
				),
			},
		},
	})
}

func TestAccStaticDNSDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSDataSourceConfig_byID("tf-acc-test-static-dns-ds-id.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_static_dns.test", "key", "tf-acc-test-static-dns-ds-id.example.com"),
					resource.TestCheckResourceAttrSet("data.unifi_static_dns.test", "id"),
				),
			},
		},
	})
}

func testAccStaticDNSDataSourceConfig_byKey(key string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = %q
  value       = "192.168.1.100"
  record_type = "A"
}

data "unifi_static_dns" "test" {
  key = unifi_static_dns.test.key
}
`, testAccProviderConfig, key)
}

func testAccStaticDNSDataSourceConfig_byID(key string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = %q
  value       = "192.168.1.101"
  record_type = "A"
}

data "unifi_static_dns" "test" {
  id = unifi_static_dns.test.id
}
`, testAccProviderConfig, key)
}
