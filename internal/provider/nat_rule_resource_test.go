package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNatRuleResource_masquerade(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_masquerade("tf-acc-test-nat-masq"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "type", "MASQUERADE"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "description", "tf-acc-test-nat-masq"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_nat_rule.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_dnat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_dnat("tf-acc-test-nat-dnat"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "type", "DNAT"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "dest_port", "80"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "translated_ip", "192.168.1.100"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "translated_port", "8080"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_snat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_snat("tf-acc-test-nat-snat"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "type", "SNAT"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "source_address", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "translated_ip", "10.0.0.1"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_withPorts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_withPorts("tf-acc-test-nat-ports"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "type", "DNAT"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "protocol", "tcp_udp"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "source_port", "1024-65535"),
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "dest_port", "443"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_disabled("tf-acc-test-nat-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_logging(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_logging("tf-acc-test-nat-logging"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "logging", "true"),
				),
			},
			{
				ResourceName:      "unifi_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNatRuleResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNatRuleResourceConfig_dnat("tf-acc-test-nat-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "translated_ip", "192.168.1.100"),
				),
			},
			{
				Config: testAccNatRuleResourceConfig_dnatUpdated("tf-acc-test-nat-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_nat_rule.test", "translated_ip", "192.168.1.200"),
				),
			},
		},
	})
}

func testAccNatRuleResourceConfig_masquerade(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type           = "MASQUERADE"
  description    = %q
  source_address = "192.168.1.0/24"
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_dnat(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type            = "DNAT"
  description     = %q
  protocol        = "tcp"
  dest_port       = "80"
  translated_ip   = "192.168.1.100"
  translated_port = "8080"
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_dnatUpdated(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type            = "DNAT"
  description     = %q
  protocol        = "tcp"
  dest_port       = "80"
  translated_ip   = "192.168.1.200"
  translated_port = "8080"
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_snat(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type           = "SNAT"
  description    = %q
  source_address = "192.168.1.0/24"
  translated_ip  = "10.0.0.1"
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_withPorts(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type            = "DNAT"
  description     = %q
  protocol        = "tcp_udp"
  source_port     = "1024-65535"
  dest_port       = "443"
  translated_ip   = "192.168.1.100"
  translated_port = "443"
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_disabled(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
  enabled     = false
}
`, testAccProviderConfig, description)
}

func testAccNatRuleResourceConfig_logging(description string) string {
	return fmt.Sprintf(`
%s

resource "unifi_nat_rule" "test" {
  type        = "MASQUERADE"
  description = %q
  logging     = true
}
`, testAccProviderConfig, description)
}
