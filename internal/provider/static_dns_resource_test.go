package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStaticDNSResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_basic("tf-acc-test-dns-basic", "192.168.1.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "key", "tf-acc-test-dns-basic.local"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "192.168.1.100"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "record_type", "A"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_static_dns.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_static_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticDNSResource_cname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_cname("tf-acc-test-dns-cname", "target.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "key", "tf-acc-test-dns-cname.local"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "target.example.com"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "record_type", "CNAME"),
				),
			},
			{
				ResourceName:      "unifi_static_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticDNSResource_mx(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_mx("tf-acc-test-dns-mx", "mail.example.com", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "key", "tf-acc-test-dns-mx.local"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "mail.example.com"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "record_type", "MX"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "priority", "10"),
				),
			},
			{
				ResourceName:      "unifi_static_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticDNSResource_srv(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_srv("tf-acc-test-dns-srv", "sip.example.com", 5060, 10, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "key", "_sip._tcp.tf-acc-test-dns-srv.local"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "sip.example.com"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "record_type", "SRV"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "port", "5060"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "priority", "10"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "weight", "5"),
				),
			},
			{
				ResourceName:      "unifi_static_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticDNSResource_withTTL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_withTTL("tf-acc-test-dns-ttl", "192.168.1.101", 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "key", "tf-acc-test-dns-ttl.local"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "192.168.1.101"),
					resource.TestCheckResourceAttr("unifi_static_dns.test", "ttl", "300"),
				),
			},
			{
				ResourceName:      "unifi_static_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStaticDNSResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_basic("tf-acc-test-dns-update", "192.168.1.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "192.168.1.100"),
				),
			},
			{
				Config: testAccStaticDNSResourceConfig_basic("tf-acc-test-dns-update", "192.168.1.200"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "value", "192.168.1.200"),
				),
			},
		},
	})
}

func TestAccStaticDNSResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticDNSResourceConfig_disabled("tf-acc-test-dns-disabled", "192.168.1.102"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_static_dns.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_static_dns.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStaticDNSResourceConfig_basic(name, value string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = "%s.local"
  value       = %q
  record_type = "A"
}
`, testAccProviderConfig, name, value)
}

func testAccStaticDNSResourceConfig_cname(name, target string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = "%s.local"
  value       = %q
  record_type = "CNAME"
}
`, testAccProviderConfig, name, target)
}

func testAccStaticDNSResourceConfig_mx(name, mailserver string, priority int) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = "%s.local"
  value       = %q
  record_type = "MX"
  priority    = %d
}
`, testAccProviderConfig, name, mailserver, priority)
}

func testAccStaticDNSResourceConfig_srv(name, target string, port, priority, weight int) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = "_sip._tcp.%s.local"
  value       = %q
  record_type = "SRV"
  port        = %d
  priority    = %d
  weight      = %d
}
`, testAccProviderConfig, name, target, port, priority, weight)
}

func testAccStaticDNSResourceConfig_withTTL(name, value string, ttl int) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = "%s.local"
  value       = %q
  record_type = "A"
  ttl         = %d
}
`, testAccProviderConfig, name, value, ttl)
}

func testAccStaticDNSResourceConfig_disabled(name, value string) string {
	return fmt.Sprintf(`
%s

resource "unifi_static_dns" "test" {
  key         = "%s.local"
  value       = %q
  record_type = "A"
  enabled     = false
}
`, testAccProviderConfig, name, value)
}
