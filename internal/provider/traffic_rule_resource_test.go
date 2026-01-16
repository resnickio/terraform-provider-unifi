package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTrafficRuleResource_block(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_block("tf-acc-test-traffic-rule-block"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-block"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "action", "BLOCK"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_traffic_rule.test", "id"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_allow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_allow("tf-acc-test-traffic-rule-allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-allow"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "action", "ALLOW"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_withSchedule(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_withSchedule("tf-acc-test-traffic-rule-schedule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-schedule"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "schedule.mode", "CUSTOM"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "schedule.time_range_start", "09:00"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "schedule.time_range_end", "17:00"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_withDomains(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_withDomains("tf-acc-test-traffic-rule-domains"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-domains"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "matching_target", "DOMAIN"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "domains.#", "2"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_withBandwidthLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_withBandwidthLimit("tf-acc-test-traffic-rule-bw"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-bw"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "bandwidth_limit.download_limit_kbps", "10000"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "bandwidth_limit.upload_limit_kbps", "5000"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "bandwidth_limit.enabled", "true"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_withIPAddresses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_withIPAddresses("tf-acc-test-traffic-rule-ips"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "name", "tf-acc-test-traffic-rule-ips"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "matching_target", "IP"),
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "ip_addresses.#", "2"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_disabled("tf-acc-test-traffic-rule-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "unifi_traffic_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTrafficRuleResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficRuleResourceConfig_block("tf-acc-test-traffic-rule-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "action", "BLOCK"),
				),
			},
			{
				Config: testAccTrafficRuleResourceConfig_allow("tf-acc-test-traffic-rule-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_traffic_rule.test", "action", "ALLOW"),
				),
			},
		},
	})
}

func testAccTrafficRuleResourceConfig_block(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name        = %q
  action      = "BLOCK"
  description = "Test block rule"
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleResourceConfig_allow(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name        = %q
  action      = "ALLOW"
  description = "Test allow rule"
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleResourceConfig_withSchedule(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name   = %q
  action = "BLOCK"

  schedule {
    mode             = "CUSTOM"
    time_range_start = "09:00"
    time_range_end   = "17:00"
    days_of_week     = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
  }
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleResourceConfig_withDomains(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name            = %q
  action          = "BLOCK"
  matching_target = "DOMAIN"

  domains {
    domain      = "*.blocked.com"
    description = "Blocked domain"
  }

  domains {
    domain = "example.blocked.com"
  }
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleResourceConfig_withBandwidthLimit(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name   = %q
  action = "ALLOW"

  bandwidth_limit {
    download_limit_kbps = 10000
    upload_limit_kbps   = 5000
    enabled             = true
  }
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleResourceConfig_withIPAddresses(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name            = %q
  action          = "BLOCK"
  matching_target = "IP"
  ip_addresses    = ["192.168.100.0/24", "10.10.10.0/24"]
}
`, testAccProviderConfig, name)
}

func testAccTrafficRuleResourceConfig_disabled(name string) string {
	return fmt.Sprintf(`
%s

resource "unifi_traffic_rule" "test" {
  name    = %q
  action  = "BLOCK"
  enabled = false
}
`, testAccProviderConfig, name)
}
