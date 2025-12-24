package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortForwardResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccPortForwardResourceConfig_basic("tf-acc-test-pf", "8443", "443", "10.0.0.50"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "dst_port", "8443"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_port", "443"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.50"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_port_forward.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_port_forward.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_udp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create UDP port forward
			{
				Config: testAccPortForwardResourceConfig_udp("tf-acc-test-pf-udp", "51820", "51820", "10.0.0.100"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-udp"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "protocol", "udp"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "dst_port", "51820"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_port", "51820"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.100"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_tcpUdp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create TCP+UDP port forward
			{
				Config: testAccPortForwardResourceConfig_tcpUdp("tf-acc-test-pf-tcpudp", "53", "53", "10.0.0.53"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-tcpudp"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "protocol", "tcp_udp"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "dst_port", "53"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_port", "53"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.53"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccPortForwardResourceConfig_basic("tf-acc-test-pf-update", "9443", "443", "10.0.0.51"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-update"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.51"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "enabled", "true"),
				),
			},
			// Update - change IP and disable
			{
				Config: testAccPortForwardResourceConfig_disabled("tf-acc-test-pf-update", "9443", "443", "10.0.0.52"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.52"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccPortForwardResource_defaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal config to verify defaults
			{
				Config: testAccPortForwardResourceConfig_minimal("tf-acc-test-pf-defaults", "8080", "80", "10.0.0.60"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-defaults"),
					// Verify defaults are applied
					resource.TestCheckResourceAttr("unifi_port_forward.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "pfwd_interface", "wan"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "log", "false"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "src", ""),
				),
			},
		},
	})
}

func TestAccPortForwardResource_withSourceRestriction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with source IP restriction
			{
				Config: testAccPortForwardResourceConfig_withSource("tf-acc-test-pf-src", "2222", "22", "10.0.0.70", "203.0.113.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-src"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "src", "203.0.113.0/24"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "dst_port", "2222"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.70"),
				),
			},
			// Update - change source restriction
			{
				Config: testAccPortForwardResourceConfig_withSource("tf-acc-test-pf-src", "2222", "22", "10.0.0.70", "198.51.100.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "src", "198.51.100.0/24"),
				),
			},
			// Update - change source restriction to different CIDR
			{
				Config: testAccPortForwardResourceConfig_withSource("tf-acc-test-pf-src", "2222", "22", "10.0.0.70", "10.0.0.0/8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "src", "10.0.0.0/8"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_withLogging(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with logging enabled
			{
				Config: testAccPortForwardResourceConfig_withLogging("tf-acc-test-pf-log", "3389", "3389", "10.0.0.80"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-log"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "log", "true"),
				),
			},
			// Update - disable logging
			{
				Config: testAccPortForwardResourceConfig_basic("tf-acc-test-pf-log", "3389", "3389", "10.0.0.80"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "log", "false"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_wan2Interface(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create on wan2 interface
			{
				Config: testAccPortForwardResourceConfig_interface("tf-acc-test-pf-wan2", "8080", "80", "10.0.0.90", "wan2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-wan2"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "pfwd_interface", "wan2"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_bothInterfaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create on both WAN interfaces
			{
				Config: testAccPortForwardResourceConfig_interface("tf-acc-test-pf-both", "9000", "9000", "10.0.0.91", "both"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-both"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "pfwd_interface", "both"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_portRange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with port range
			{
				Config: testAccPortForwardResourceConfig_basic("tf-acc-test-pf-range", "30000-30010", "30000-30010", "10.0.0.92"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-range"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "dst_port", "30000-30010"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_port", "30000-30010"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortForwardResource_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all options
			{
				Config: testAccPortForwardResourceConfig_full("tf-acc-test-pf-full", "4443", "443", "10.0.0.93", "192.0.2.0/24", "wan", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_forward.test", "name", "tf-acc-test-pf-full"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "dst_port", "4443"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_port", "443"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "fwd_ip", "10.0.0.93"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "src", "192.0.2.0/24"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "pfwd_interface", "wan"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "log", "true"),
					resource.TestCheckResourceAttr("unifi_port_forward.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("unifi_port_forward.test", "id"),
					resource.TestCheckResourceAttrSet("unifi_port_forward.test", "site_id"),
				),
			},
			// ImportState
			{
				ResourceName:      "unifi_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPortForwardResourceConfig_basic(name, dstPort, fwdPort, fwdIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
  enabled  = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP)
}

func testAccPortForwardResourceConfig_udp(name, dstPort, fwdPort, fwdIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "udp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
  enabled  = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP)
}

func testAccPortForwardResourceConfig_tcpUdp(name, dstPort, fwdPort, fwdIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp_udp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
  enabled  = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP)
}

func testAccPortForwardResourceConfig_disabled(name, dstPort, fwdPort, fwdIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
  enabled  = false
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP)
}

func testAccPortForwardResourceConfig_minimal(name, dstPort, fwdPort, fwdIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP)
}

func testAccPortForwardResourceConfig_withSource(name, dstPort, fwdPort, fwdIP, src string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
  src      = %q
  enabled  = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP, src)
}

func testAccPortForwardResourceConfig_withLogging(name, dstPort, fwdPort, fwdIP string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name     = %q
  protocol = "tcp"
  dst_port = %q
  fwd_port = %q
  fwd_ip   = %q
  log      = true
  enabled  = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP)
}

func testAccPortForwardResourceConfig_interface(name, dstPort, fwdPort, fwdIP, iface string) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name           = %q
  protocol       = "tcp"
  dst_port       = %q
  fwd_port       = %q
  fwd_ip         = %q
  pfwd_interface = %q
  enabled        = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP, iface)
}

func testAccPortForwardResourceConfig_full(name, dstPort, fwdPort, fwdIP, src, iface string, log bool) string {
	return fmt.Sprintf(`
%s

resource "unifi_port_forward" "test" {
  name           = %q
  protocol       = "tcp"
  dst_port       = %q
  fwd_port       = %q
  fwd_ip         = %q
  src            = %q
  pfwd_interface = %q
  log            = %t
  enabled        = true
}
`, testAccProviderConfig, name, dstPort, fwdPort, fwdIP, src, iface, log)
}
