package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

// formatStringListForHCL formats a slice of strings as an HCL list literal.
func formatStringListForHCL(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%q", item)
	}
	return result
}

// testAccGetClient creates an SDK client for test setup/teardown operations.
func testAccGetClient(t *testing.T) *unifi.NetworkClient {
	config := unifi.NetworkClientConfig{
		BaseURL:            os.Getenv("UNIFI_BASE_URL"),
		APIKey:             os.Getenv("UNIFI_API_KEY"),
		Site:               "default",
		InsecureSkipVerify: os.Getenv("UNIFI_INSECURE") == "true",
	}

	client, err := unifi.NewNetworkClient(config)
	if err != nil {
		t.Skipf("Could not create SDK client: %v", err)
		return nil
	}

	return client
}

// testAccCheckControllerSupportsZones checks if the controller supports zone-based firewall.
func testAccCheckControllerSupportsZones(t *testing.T) {
	testAccPreCheck(t)

	client := testAccGetClient(t)
	if client == nil {
		return
	}

	_, err := client.CreateFirewallZone(context.Background(), &unifi.FirewallZoneCreateRequest{
		Name:       "tf-acc-zone-test-precheck",
		NetworkIDs: []string{},
	})
	if err != nil {
		t.Skipf("Controller does not support firewall zones: %v", err)
		return
	}

	zones, err := client.ListFirewallZones(context.Background())
	if err == nil {
		for _, zone := range zones {
			if zone.Name == "tf-acc-zone-test-precheck" {
				_ = client.DeleteFirewallZone(context.Background(), zone.ID)
				break
			}
		}
	}
}

// testAccCheckControllerSupportsGuestNetworks checks if the controller supports guest networks.
// Some controllers (like UDM) may convert guest networks to corporate.
func testAccCheckControllerSupportsGuestNetworks(t *testing.T) {
	testAccPreCheck(t)

	client := testAccGetClient(t)
	if client == nil {
		return
	}

	ctx := context.Background()

	network := &unifi.Network{
		Name:    "tf-acc-guest-test-precheck",
		Purpose: "guest",
	}

	created, err := client.CreateNetwork(ctx, network)
	if err != nil {
		t.Skipf("Controller does not support guest network creation: %v", err)
		return
	}

	defer func() {
		_ = client.DeleteNetwork(ctx, created.ID)
	}()

	if created.Purpose != "guest" {
		t.Skipf("Controller does not support guest networks: purpose was changed from 'guest' to '%s'", created.Purpose)
	}
}

// testAccCheckControllerSupportsLegacyRules checks if controller supports legacy firewall rules.
// Controllers with zone-based firewall (UDM Network 10.x) don't support legacy rules.
func testAccCheckControllerSupportsLegacyRules(t *testing.T) {
	testAccPreCheck(t)

	client := testAccGetClient(t)
	if client == nil {
		return
	}

	zones, err := client.ListFirewallZones(context.Background())
	if err == nil && len(zones) > 0 {
		t.Skip("Controller uses zone-based firewall, legacy rules not supported")
	}
}

// testAccPreCheckDevice checks if there is at least one adopted device available for testing.
func testAccPreCheckDevice(t *testing.T) {
	testAccPreCheck(t)

	client := testAccGetClient(t)
	if client == nil {
		return
	}

	devices, err := client.ListDevices(context.Background())
	if err != nil {
		t.Skipf("Could not list devices: %v", err)
		return
	}

	if devices == nil || len(devices.NetworkDevices) == 0 {
		t.Skip("No adopted devices available for testing")
		return
	}
}

// testAccPreCheckSwitch checks if there is at least one switch device available for testing.
func testAccPreCheckSwitch(t *testing.T) {
	testAccPreCheck(t)

	client := testAccGetClient(t)
	if client == nil {
		return
	}

	devices, err := client.ListDevices(context.Background())
	if err != nil {
		t.Skipf("Could not list devices: %v", err)
		return
	}

	if devices == nil {
		t.Skip("No devices available for testing")
		return
	}

	for _, d := range devices.NetworkDevices {
		if d.Type == "usw" {
			return
		}
	}
	t.Skip("No switch devices available for testing")
}

// testAccGetFirstDeviceMAC returns the MAC of the first adopted device.
func testAccGetFirstDeviceMAC(t *testing.T) string {
	client := testAccGetClient(t)
	if client == nil {
		return ""
	}

	devices, err := client.ListDevices(context.Background())
	if err != nil || devices == nil || len(devices.NetworkDevices) == 0 {
		return ""
	}

	return devices.NetworkDevices[0].MAC
}

// testAccGetFirstSwitchMAC returns the MAC of the first switch device.
func testAccGetFirstSwitchMAC(t *testing.T) string {
	client := testAccGetClient(t)
	if client == nil {
		return ""
	}

	devices, err := client.ListDevices(context.Background())
	if err != nil || devices == nil {
		return ""
	}

	for _, d := range devices.NetworkDevices {
		if d.Type == "usw" {
			return d.MAC
		}
	}
	return ""
}

// testAccFirewallZonePairConfig returns configuration for a pair of firewall zones.
func testAccFirewallZonePairConfig(baseName string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone" "src" {
  name = "%s-src-zone"
}

resource "unifi_firewall_zone" "dst" {
  name = "%s-dst-zone"
}
`, baseName, baseName)
}
