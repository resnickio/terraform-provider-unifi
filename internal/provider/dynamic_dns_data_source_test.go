package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDynamicDNSDataSource_byHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSDataSourceConfig_byHostname("tf-acc-test-ddns-ds.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-ds.example.com"),
					resource.TestCheckResourceAttr("data.unifi_dynamic_dns.test", "service", "custom"),
					resource.TestCheckResourceAttrSet("data.unifi_dynamic_dns.test", "id"),
				),
			},
		},
	})
}

func TestAccDynamicDNSDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSDataSourceConfig_byID("tf-acc-test-ddns-ds-id.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-ds-id.example.com"),
					resource.TestCheckResourceAttrSet("data.unifi_dynamic_dns.test", "id"),
				),
			},
		},
	})
}

func testAccDynamicDNSDataSourceConfig_byHostname(hostname string) string {
	return fmt.Sprintf(`
%s

resource "unifi_dynamic_dns" "test" {
  service  = "custom"
  hostname = %q
  server   = "update.example.com"
  login    = "test@example.com"
  password = "test-api-token"
}

data "unifi_dynamic_dns" "test" {
  hostname = unifi_dynamic_dns.test.hostname
}
`, testAccProviderConfig, hostname)
}

func testAccDynamicDNSDataSourceConfig_byID(hostname string) string {
	return fmt.Sprintf(`
%s

resource "unifi_dynamic_dns" "test" {
  service  = "custom"
  hostname = %q
  server   = "update.example.com"
  login    = "test@example.com"
  password = "test-api-token"
}

data "unifi_dynamic_dns" "test" {
  id = unifi_dynamic_dns.test.id
}
`, testAccProviderConfig, hostname)
}
