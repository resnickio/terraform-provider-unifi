package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during
// acceptance testing. The factory function is called for every Terraform
// CLI command executed to create a provider server to which the CLI can
// connect.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unifi": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates the necessary environment variables are set
// for running acceptance tests.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	if v := os.Getenv("UNIFI_BASE_URL"); v == "" {
		t.Fatal("UNIFI_BASE_URL must be set for acceptance tests")
	}
	// Either API key or username/password must be set
	apiKey := os.Getenv("UNIFI_API_KEY")
	username := os.Getenv("UNIFI_USERNAME")
	password := os.Getenv("UNIFI_PASSWORD")

	if apiKey == "" && (username == "" || password == "") {
		t.Fatal("Either UNIFI_API_KEY or both UNIFI_USERNAME and UNIFI_PASSWORD must be set for acceptance tests")
	}
}

// testAccProviderConfig is the base provider configuration for tests.
// It uses environment variables for all configuration values.
const testAccProviderConfig = `
provider "unifi" {}
`
