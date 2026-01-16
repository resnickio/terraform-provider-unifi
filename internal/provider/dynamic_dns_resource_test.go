package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDynamicDNSResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSResourceConfig_basic("tf-acc-test-ddns-basic.duckdns.org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-basic.duckdns.org"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "service", "duckdns"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "interface", "wan"),
					resource.TestCheckResourceAttrSet("unifi_dynamic_dns.test", "id"),
				),
			},
			{
				ResourceName:            "unifi_dynamic_dns.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDynamicDNSResource_cloudflare(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSResourceConfig_cloudflare("tf-acc-test-ddns-cf.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-cf.example.com"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "service", "cloudflare"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "login", "test@example.com"),
				),
			},
			{
				ResourceName:            "unifi_dynamic_dns.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDynamicDNSResource_custom(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSResourceConfig_custom("tf-acc-test-ddns-custom.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-custom.example.com"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "service", "custom"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "server", "update.example.com"),
				),
			},
			{
				ResourceName:            "unifi_dynamic_dns.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDynamicDNSResource_wan2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSResourceConfig_wan2("tf-acc-test-ddns-wan2.duckdns.org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-wan2.duckdns.org"),
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "interface", "wan2"),
				),
			},
			{
				ResourceName:            "unifi_dynamic_dns.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDynamicDNSResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSResourceConfig_basic("tf-acc-test-ddns-update.duckdns.org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-update.duckdns.org"),
				),
			},
			{
				Config: testAccDynamicDNSResourceConfig_basic("tf-acc-test-ddns-updated.duckdns.org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_dynamic_dns.test", "hostname", "tf-acc-test-ddns-updated.duckdns.org"),
				),
			},
			{
				ResourceName:            "unifi_dynamic_dns.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccDynamicDNSResourceConfig_basic(hostname string) string {
	return fmt.Sprintf(`
%s

resource "unifi_dynamic_dns" "test" {
  service  = "duckdns"
  hostname = %q
  password = "test-token"
}
`, testAccProviderConfig, hostname)
}

func testAccDynamicDNSResourceConfig_cloudflare(hostname string) string {
	return fmt.Sprintf(`
%s

resource "unifi_dynamic_dns" "test" {
  service  = "cloudflare"
  hostname = %q
  login    = "test@example.com"
  password = "test-api-token"
}
`, testAccProviderConfig, hostname)
}

func testAccDynamicDNSResourceConfig_custom(hostname string) string {
	return fmt.Sprintf(`
%s

resource "unifi_dynamic_dns" "test" {
  service  = "custom"
  hostname = %q
  server   = "update.example.com"
  login    = "testuser"
  password = "testpass"
}
`, testAccProviderConfig, hostname)
}

func testAccDynamicDNSResourceConfig_wan2(hostname string) string {
	return fmt.Sprintf(`
%s

resource "unifi_dynamic_dns" "test" {
  service   = "duckdns"
  hostname  = %q
  password  = "test-token"
  interface = "wan2"
}
`, testAccProviderConfig, hostname)
}
