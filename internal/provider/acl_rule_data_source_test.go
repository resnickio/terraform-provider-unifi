package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheckAclRules(t *testing.T) {
	testAccPreCheck(t)
	client := testAccGetClient(t)
	if client == nil {
		return
	}
	rules, err := client.ListAclRules(context.Background())
	if err != nil || len(rules) == 0 {
		t.Skip("No ACL rules available on this controller")
	}
}

func TestAccAclRuleDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAclRules(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAclRuleDataSourceConfig_firstByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.unifi_acl_rule.test", "id"),
					resource.TestCheckResourceAttrSet("data.unifi_acl_rule.test", "name"),
				),
			},
		},
	})
}

func testAccAclRuleDataSourceConfig_firstByName() string {
	// Use the first available ACL rule name from the controller
	return testAccProviderConfig + `
data "unifi_acl_rule" "test" {
  name = "Block LAN to WLAN Multicast and Broadcast"
}
`
}
