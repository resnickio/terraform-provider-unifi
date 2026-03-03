package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckQosRules(t *testing.T) {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		return
	}
	rules, err := client.ListQosRules(context.Background())
	if err != nil || len(rules) == 0 {
		t.Skip("No QoS rules available on this controller")
	}
}

func TestAccQosRuleDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckQosRules(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQosRuleDataSourceConfig_byName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_qos_rule.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_qos_rule.test", "name"),
				),
			},
		},
	})
}

func testAccQosRuleDataSourceConfig_byName() string {
	return testAccProviderConfig + `
data "unifi_qos_rule" "test" {
  name = "Default"
}
`
}
