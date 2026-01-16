package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

const testResourcePrefix = "tf-acc-test-"

func init() {
	resource.AddTestSweepers("unifi_network", &resource.Sweeper{
		Name: "unifi_network",
		F:    sweepNetworks,
	})

	resource.AddTestSweepers("unifi_firewall_group", &resource.Sweeper{
		Name: "unifi_firewall_group",
		F:    sweepFirewallGroups,
	})

	resource.AddTestSweepers("unifi_firewall_rule", &resource.Sweeper{
		Name:         "unifi_firewall_rule",
		F:            sweepFirewallRules,
		Dependencies: []string{"unifi_firewall_group"},
	})

	resource.AddTestSweepers("unifi_firewall_policy", &resource.Sweeper{
		Name:         "unifi_firewall_policy",
		F:            sweepFirewallPolicies,
		Dependencies: []string{"unifi_firewall_zone"},
	})

	resource.AddTestSweepers("unifi_firewall_zone", &resource.Sweeper{
		Name:         "unifi_firewall_zone",
		F:            sweepFirewallZones,
		Dependencies: []string{"unifi_network"},
	})

	resource.AddTestSweepers("unifi_port_forward", &resource.Sweeper{
		Name: "unifi_port_forward",
		F:    sweepPortForwards,
	})

	resource.AddTestSweepers("unifi_static_route", &resource.Sweeper{
		Name: "unifi_static_route",
		F:    sweepStaticRoutes,
	})

	resource.AddTestSweepers("unifi_user_group", &resource.Sweeper{
		Name: "unifi_user_group",
		F:    sweepUserGroups,
	})

	resource.AddTestSweepers("unifi_wlan", &resource.Sweeper{
		Name:         "unifi_wlan",
		F:            sweepWLANs,
		Dependencies: []string{"unifi_user_group"},
	})

	resource.AddTestSweepers("unifi_port_profile", &resource.Sweeper{
		Name:         "unifi_port_profile",
		F:            sweepPortProfiles,
		Dependencies: []string{"unifi_network"},
	})

	resource.AddTestSweepers("unifi_static_dns", &resource.Sweeper{
		Name: "unifi_static_dns",
		F:    sweepStaticDNS,
	})

	resource.AddTestSweepers("unifi_dynamic_dns", &resource.Sweeper{
		Name: "unifi_dynamic_dns",
		F:    sweepDynamicDNS,
	})

	resource.AddTestSweepers("unifi_nat_rule", &resource.Sweeper{
		Name: "unifi_nat_rule",
		F:    sweepNatRules,
	})

	resource.AddTestSweepers("unifi_traffic_rule", &resource.Sweeper{
		Name: "unifi_traffic_rule",
		F:    sweepTrafficRules,
	})

	resource.AddTestSweepers("unifi_traffic_route", &resource.Sweeper{
		Name: "unifi_traffic_route",
		F:    sweepTrafficRoutes,
	})

	resource.AddTestSweepers("unifi_radius_profile", &resource.Sweeper{
		Name: "unifi_radius_profile",
		F:    sweepRADIUSProfiles,
	})
}

func getSweeperClient() (*unifi.NetworkClient, error) {
	config := unifi.NetworkClientConfig{
		BaseURL:            os.Getenv("UNIFI_BASE_URL"),
		Site:               os.Getenv("UNIFI_SITE"),
		InsecureSkipVerify: os.Getenv("UNIFI_INSECURE") == "true",
	}

	if config.Site == "" {
		config.Site = "default"
	}

	if apiKey := os.Getenv("UNIFI_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	} else {
		config.Username = os.Getenv("UNIFI_USERNAME")
		config.Password = os.Getenv("UNIFI_PASSWORD")
	}

	client, err := unifi.NewNetworkClient(config)
	if err != nil {
		return nil, err
	}

	if config.APIKey == "" {
		if err := client.Login(context.Background()); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func sweepNetworks(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	networks, err := client.ListNetworks(ctx)
	if err != nil {
		return err
	}

	for _, network := range networks {
		if strings.HasPrefix(network.Name, testResourcePrefix) {
			if err := client.DeleteNetwork(ctx, network.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepFirewallGroups(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	groups, err := client.ListFirewallGroups(ctx)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if strings.HasPrefix(group.Name, testResourcePrefix) {
			if err := client.DeleteFirewallGroup(ctx, group.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepFirewallRules(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	rules, err := client.ListFirewallRules(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if strings.HasPrefix(rule.Name, testResourcePrefix) {
			if err := client.DeleteFirewallRule(ctx, rule.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepFirewallPolicies(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	policies, err := client.ListFirewallPolicies(ctx)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if strings.HasPrefix(policy.Name, testResourcePrefix) {
			if err := client.DeleteFirewallPolicy(ctx, policy.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepFirewallZones(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	zones, err := client.ListFirewallZones(ctx)
	if err != nil {
		return err
	}

	for _, zone := range zones {
		if strings.HasPrefix(zone.Name, testResourcePrefix) {
			if err := client.DeleteFirewallZone(ctx, zone.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepPortForwards(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	portForwards, err := client.ListPortForwards(ctx)
	if err != nil {
		return err
	}

	for _, pf := range portForwards {
		if strings.HasPrefix(pf.Name, testResourcePrefix) {
			if err := client.DeletePortForward(ctx, pf.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepStaticRoutes(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	routes, err := client.ListRoutes(ctx)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if strings.HasPrefix(route.Name, testResourcePrefix) {
			if err := client.DeleteRoute(ctx, route.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepUserGroups(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	groups, err := client.ListUserGroups(ctx)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if strings.HasPrefix(group.Name, testResourcePrefix) {
			if err := client.DeleteUserGroup(ctx, group.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepWLANs(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	wlans, err := client.ListWLANs(ctx)
	if err != nil {
		return err
	}

	for _, wlan := range wlans {
		if strings.HasPrefix(wlan.Name, testResourcePrefix) {
			if err := client.DeleteWLAN(ctx, wlan.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepPortProfiles(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	profiles, err := client.ListPortConfs(ctx)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		if strings.HasPrefix(profile.Name, testResourcePrefix) {
			if err := client.DeletePortConf(ctx, profile.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepStaticDNS(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	records, err := client.ListStaticDNS(ctx)
	if err != nil {
		return err
	}

	for _, record := range records {
		if strings.HasPrefix(record.Key, testResourcePrefix) {
			if err := client.DeleteStaticDNS(ctx, record.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepDynamicDNS(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	records, err := client.ListDynamicDNS(ctx)
	if err != nil {
		return err
	}

	for _, record := range records {
		if strings.HasPrefix(record.HostName, testResourcePrefix) {
			if err := client.DeleteDynamicDNS(ctx, record.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepNatRules(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	rules, err := client.ListNatRules(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if strings.HasPrefix(rule.Description, testResourcePrefix) {
			if err := client.DeleteNatRule(ctx, rule.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepTrafficRules(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	rules, err := client.ListTrafficRules(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if strings.HasPrefix(rule.Name, testResourcePrefix) {
			if err := client.DeleteTrafficRule(ctx, rule.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepTrafficRoutes(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	routes, err := client.ListTrafficRoutes(ctx)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if strings.HasPrefix(route.Name, testResourcePrefix) {
			if err := client.DeleteTrafficRoute(ctx, route.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func sweepRADIUSProfiles(region string) error {
	client, err := getSweeperClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	profiles, err := client.ListRADIUSProfiles(ctx)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		if strings.HasPrefix(profile.Name, testResourcePrefix) {
			if err := client.DeleteRADIUSProfile(ctx, profile.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
