package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

const minAuthInterval = 5 * time.Second

// AutoLoginClient wraps a NetworkManager to automatically re-authenticate on session expiration.
type AutoLoginClient struct {
	client       unifi.NetworkManager
	config       unifi.NetworkClientConfig
	mu           sync.Mutex
	lastAuthTime time.Time
	authSem      chan struct{}
	deviceMu     sync.Map // map[string]*sync.Mutex for per-device locking
}

// NewAutoLoginClient creates a new auto-login wrapper around the SDK client.
func NewAutoLoginClient(client unifi.NetworkManager, config unifi.NetworkClientConfig) *AutoLoginClient {
	return &AutoLoginClient{
		client:  client,
		config:  config,
		authSem: make(chan struct{}, 1),
	}
}

// withRetry executes the given function and retries with re-authentication if unauthorized.
func (c *AutoLoginClient) withRetry(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil || !errors.Is(err, unifi.ErrUnauthorized) {
		return err
	}

	// Record when this request failed
	failedAt := time.Now()

	// Try to acquire the semaphore for re-authentication
	select {
	case c.authSem <- struct{}{}:
		// We got the semaphore, we'll handle re-auth
		defer func() { <-c.authSem }()
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(minAuthInterval):
		// Another goroutine is handling auth or we waited long enough, just retry
		return fn()
	}

	c.mu.Lock()

	// Double-check: if we already re-authenticated after this request started,
	// just retry without re-authenticating again
	if c.lastAuthTime.After(failedAt) {
		c.mu.Unlock()
		return fn()
	}

	// Rate limit: context-aware wait if needed
	if timeSinceLastAuth := time.Since(c.lastAuthTime); timeSinceLastAuth < minAuthInterval {
		waitTime := minAuthInterval - timeSinceLastAuth
		c.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}

		c.mu.Lock()
	}

	// Re-authenticate
	loginErr := c.client.Login(ctx)
	if loginErr != nil {
		c.mu.Unlock()
		return fmt.Errorf("re-authentication failed: %w", loginErr)
	}
	c.lastAuthTime = time.Now()
	c.mu.Unlock()

	// Retry the operation
	return fn()
}

// Network operations

func (c *AutoLoginClient) ListNetworks(ctx context.Context) ([]unifi.Network, error) {
	var result []unifi.Network
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListNetworks(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetNetwork(ctx context.Context, id string) (*unifi.Network, error) {
	var result *unifi.Network
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetNetwork(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateNetwork(ctx context.Context, network *unifi.Network) (*unifi.Network, error) {
	var result *unifi.Network
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateNetwork(ctx, network)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateNetwork(ctx context.Context, id string, network *unifi.Network) (*unifi.Network, error) {
	var result *unifi.Network
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateNetwork(ctx, id, network)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteNetwork(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteNetwork(ctx, id)
	})
}

// Firewall Rule operations

func (c *AutoLoginClient) ListFirewallRules(ctx context.Context) ([]unifi.FirewallRule, error) {
	var result []unifi.FirewallRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListFirewallRules(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetFirewallRule(ctx context.Context, id string) (*unifi.FirewallRule, error) {
	var result *unifi.FirewallRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetFirewallRule(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateFirewallRule(ctx context.Context, rule *unifi.FirewallRule) (*unifi.FirewallRule, error) {
	var result *unifi.FirewallRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateFirewallRule(ctx, rule)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateFirewallRule(ctx context.Context, id string, rule *unifi.FirewallRule) (*unifi.FirewallRule, error) {
	var result *unifi.FirewallRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateFirewallRule(ctx, id, rule)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteFirewallRule(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteFirewallRule(ctx, id)
	})
}

// Firewall Group operations

func (c *AutoLoginClient) ListFirewallGroups(ctx context.Context) ([]unifi.FirewallGroup, error) {
	var result []unifi.FirewallGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListFirewallGroups(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetFirewallGroup(ctx context.Context, id string) (*unifi.FirewallGroup, error) {
	var result *unifi.FirewallGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetFirewallGroup(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateFirewallGroup(ctx context.Context, group *unifi.FirewallGroup) (*unifi.FirewallGroup, error) {
	var result *unifi.FirewallGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateFirewallGroup(ctx, group)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateFirewallGroup(ctx context.Context, id string, group *unifi.FirewallGroup) (*unifi.FirewallGroup, error) {
	var result *unifi.FirewallGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateFirewallGroup(ctx, id, group)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteFirewallGroup(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteFirewallGroup(ctx, id)
	})
}

// Port Forward operations

func (c *AutoLoginClient) ListPortForwards(ctx context.Context) ([]unifi.PortForward, error) {
	var result []unifi.PortForward
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListPortForwards(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetPortForward(ctx context.Context, id string) (*unifi.PortForward, error) {
	var result *unifi.PortForward
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetPortForward(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreatePortForward(ctx context.Context, pf *unifi.PortForward) (*unifi.PortForward, error) {
	var result *unifi.PortForward
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreatePortForward(ctx, pf)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdatePortForward(ctx context.Context, id string, pf *unifi.PortForward) (*unifi.PortForward, error) {
	var result *unifi.PortForward
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdatePortForward(ctx, id, pf)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeletePortForward(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeletePortForward(ctx, id)
	})
}

// WLAN operations

func (c *AutoLoginClient) ListWLANs(ctx context.Context) ([]unifi.WLANConf, error) {
	var result []unifi.WLANConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListWLANs(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetWLAN(ctx context.Context, id string) (*unifi.WLANConf, error) {
	var result *unifi.WLANConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetWLAN(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateWLAN(ctx context.Context, wlan *unifi.WLANConf) (*unifi.WLANConf, error) {
	var result *unifi.WLANConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateWLAN(ctx, wlan)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateWLAN(ctx context.Context, id string, wlan *unifi.WLANConf) (*unifi.WLANConf, error) {
	var result *unifi.WLANConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateWLAN(ctx, id, wlan)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteWLAN(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteWLAN(ctx, id)
	})
}

// Firewall Policy operations (v2 zone-based firewall)

func (c *AutoLoginClient) ListFirewallPolicies(ctx context.Context) ([]unifi.FirewallPolicy, error) {
	var result []unifi.FirewallPolicy
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListFirewallPolicies(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetFirewallPolicy(ctx context.Context, id string) (*unifi.FirewallPolicy, error) {
	var result *unifi.FirewallPolicy
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetFirewallPolicy(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateFirewallPolicy(ctx context.Context, policy *unifi.FirewallPolicy) (*unifi.FirewallPolicy, error) {
	var result *unifi.FirewallPolicy
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateFirewallPolicy(ctx, policy)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateFirewallPolicy(ctx context.Context, id string, policy *unifi.FirewallPolicy) (*unifi.FirewallPolicy, error) {
	var result *unifi.FirewallPolicy
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateFirewallPolicy(ctx, id, policy)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteFirewallPolicy(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteFirewallPolicy(ctx, id)
	})
}

// Firewall Zone operations

func (c *AutoLoginClient) ListFirewallZones(ctx context.Context) ([]unifi.FirewallZone, error) {
	var result []unifi.FirewallZone
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListFirewallZones(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetFirewallZone(ctx context.Context, id string) (*unifi.FirewallZone, error) {
	var result *unifi.FirewallZone
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetFirewallZone(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateFirewallZone(ctx context.Context, req *unifi.FirewallZoneCreateRequest) (*unifi.FirewallZone, error) {
	var result *unifi.FirewallZone
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateFirewallZone(ctx, req)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateFirewallZone(ctx context.Context, id string, req *unifi.FirewallZoneUpdateRequest) (*unifi.FirewallZone, error) {
	var result *unifi.FirewallZone
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateFirewallZone(ctx, id, req)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteFirewallZone(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteFirewallZone(ctx, id)
	})
}

// Static Route operations

func (c *AutoLoginClient) ListRoutes(ctx context.Context) ([]unifi.Routing, error) {
	var result []unifi.Routing
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListRoutes(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetRoute(ctx context.Context, id string) (*unifi.Routing, error) {
	var result *unifi.Routing
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetRoute(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateRoute(ctx context.Context, route *unifi.Routing) (*unifi.Routing, error) {
	var result *unifi.Routing
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateRoute(ctx, route)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateRoute(ctx context.Context, id string, route *unifi.Routing) (*unifi.Routing, error) {
	var result *unifi.Routing
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateRoute(ctx, id, route)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteRoute(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteRoute(ctx, id)
	})
}

// User Group operations

func (c *AutoLoginClient) ListUserGroups(ctx context.Context) ([]unifi.UserGroup, error) {
	var result []unifi.UserGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListUserGroups(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetUserGroup(ctx context.Context, id string) (*unifi.UserGroup, error) {
	var result *unifi.UserGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetUserGroup(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateUserGroup(ctx context.Context, group *unifi.UserGroup) (*unifi.UserGroup, error) {
	var result *unifi.UserGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateUserGroup(ctx, group)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateUserGroup(ctx context.Context, id string, group *unifi.UserGroup) (*unifi.UserGroup, error) {
	var result *unifi.UserGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateUserGroup(ctx, id, group)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteUserGroup(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteUserGroup(ctx, id)
	})
}

// AP Group operations

func (c *AutoLoginClient) ListAPGroups(ctx context.Context) ([]unifi.APGroup, error) {
	var result []unifi.APGroup
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListAPGroups(ctx)
		return err
	})
	return result, err
}

// Port Profile operations

func (c *AutoLoginClient) ListPortProfiles(ctx context.Context) ([]unifi.PortConf, error) {
	var result []unifi.PortConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListPortConfs(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetPortProfile(ctx context.Context, id string) (*unifi.PortConf, error) {
	var result *unifi.PortConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetPortConf(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreatePortProfile(ctx context.Context, p *unifi.PortConf) (*unifi.PortConf, error) {
	var result *unifi.PortConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreatePortConf(ctx, p)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdatePortProfile(ctx context.Context, id string, p *unifi.PortConf) (*unifi.PortConf, error) {
	var result *unifi.PortConf
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdatePortConf(ctx, id, p)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeletePortProfile(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeletePortConf(ctx, id)
	})
}

// Static DNS operations

func (c *AutoLoginClient) ListStaticDNS(ctx context.Context) ([]unifi.StaticDNS, error) {
	var result []unifi.StaticDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListStaticDNS(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetStaticDNS(ctx context.Context, id string) (*unifi.StaticDNS, error) {
	var result *unifi.StaticDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetStaticDNS(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateStaticDNS(ctx context.Context, dns *unifi.StaticDNS) (*unifi.StaticDNS, error) {
	var result *unifi.StaticDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateStaticDNS(ctx, dns)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateStaticDNS(ctx context.Context, id string, dns *unifi.StaticDNS) (*unifi.StaticDNS, error) {
	var result *unifi.StaticDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateStaticDNS(ctx, id, dns)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteStaticDNS(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteStaticDNS(ctx, id)
	})
}

// Dynamic DNS operations

func (c *AutoLoginClient) ListDynamicDNS(ctx context.Context) ([]unifi.DynamicDNS, error) {
	var result []unifi.DynamicDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListDynamicDNS(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetDynamicDNS(ctx context.Context, id string) (*unifi.DynamicDNS, error) {
	var result *unifi.DynamicDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetDynamicDNS(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateDynamicDNS(ctx context.Context, dns *unifi.DynamicDNS) (*unifi.DynamicDNS, error) {
	var result *unifi.DynamicDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateDynamicDNS(ctx, dns)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateDynamicDNS(ctx context.Context, id string, dns *unifi.DynamicDNS) (*unifi.DynamicDNS, error) {
	var result *unifi.DynamicDNS
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateDynamicDNS(ctx, id, dns)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteDynamicDNS(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteDynamicDNS(ctx, id)
	})
}

// NAT Rule operations

func (c *AutoLoginClient) ListNatRules(ctx context.Context) ([]unifi.NatRule, error) {
	var result []unifi.NatRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListNatRules(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetNatRule(ctx context.Context, id string) (*unifi.NatRule, error) {
	var result *unifi.NatRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetNatRule(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateNatRule(ctx context.Context, rule *unifi.NatRule) (*unifi.NatRule, error) {
	var result *unifi.NatRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateNatRule(ctx, rule)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateNatRule(ctx context.Context, id string, rule *unifi.NatRule) (*unifi.NatRule, error) {
	var result *unifi.NatRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateNatRule(ctx, id, rule)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteNatRule(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteNatRule(ctx, id)
	})
}

// Traffic Rule operations

func (c *AutoLoginClient) ListTrafficRules(ctx context.Context) ([]unifi.TrafficRule, error) {
	var result []unifi.TrafficRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListTrafficRules(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetTrafficRule(ctx context.Context, id string) (*unifi.TrafficRule, error) {
	var result *unifi.TrafficRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetTrafficRule(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateTrafficRule(ctx context.Context, rule *unifi.TrafficRule) (*unifi.TrafficRule, error) {
	var result *unifi.TrafficRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateTrafficRule(ctx, rule)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateTrafficRule(ctx context.Context, id string, rule *unifi.TrafficRule) (*unifi.TrafficRule, error) {
	var result *unifi.TrafficRule
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateTrafficRule(ctx, id, rule)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteTrafficRule(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteTrafficRule(ctx, id)
	})
}

// Traffic Route operations

func (c *AutoLoginClient) ListTrafficRoutes(ctx context.Context) ([]unifi.TrafficRoute, error) {
	var result []unifi.TrafficRoute
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListTrafficRoutes(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetTrafficRoute(ctx context.Context, id string) (*unifi.TrafficRoute, error) {
	var result *unifi.TrafficRoute
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetTrafficRoute(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateTrafficRoute(ctx context.Context, route *unifi.TrafficRoute) (*unifi.TrafficRoute, error) {
	var result *unifi.TrafficRoute
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateTrafficRoute(ctx, route)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateTrafficRoute(ctx context.Context, id string, route *unifi.TrafficRoute) (*unifi.TrafficRoute, error) {
	var result *unifi.TrafficRoute
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateTrafficRoute(ctx, id, route)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteTrafficRoute(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteTrafficRoute(ctx, id)
	})
}

// RADIUS Profile operations

func (c *AutoLoginClient) ListRADIUSProfiles(ctx context.Context) ([]unifi.RADIUSProfile, error) {
	var result []unifi.RADIUSProfile
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListRADIUSProfiles(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetRADIUSProfile(ctx context.Context, id string) (*unifi.RADIUSProfile, error) {
	var result *unifi.RADIUSProfile
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetRADIUSProfile(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateRADIUSProfile(ctx context.Context, profile *unifi.RADIUSProfile) (*unifi.RADIUSProfile, error) {
	var result *unifi.RADIUSProfile
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateRADIUSProfile(ctx, profile)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateRADIUSProfile(ctx context.Context, id string, profile *unifi.RADIUSProfile) (*unifi.RADIUSProfile, error) {
	var result *unifi.RADIUSProfile
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateRADIUSProfile(ctx, id, profile)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteRADIUSProfile(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteRADIUSProfile(ctx, id)
	})
}

// Device operations

// getDeviceLock returns a mutex for the given device ID, creating one if needed.
func (c *AutoLoginClient) getDeviceLock(deviceID string) *sync.Mutex {
	actual, _ := c.deviceMu.LoadOrStore(deviceID, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (c *AutoLoginClient) ListDevices(ctx context.Context) (*unifi.DeviceList, error) {
	var result *unifi.DeviceList
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListDevices(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetDeviceByMAC(ctx context.Context, mac string) (*unifi.DeviceConfig, error) {
	var result *unifi.DeviceConfig
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetDeviceByMAC(ctx, mac)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateDevice(ctx context.Context, id string, device *unifi.DeviceConfig) (*unifi.DeviceConfig, error) {
	var result *unifi.DeviceConfig
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateDevice(ctx, id, device)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) ListUsers(ctx context.Context) ([]unifi.User, error) {
	var result []unifi.User
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.ListUsers(ctx)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) GetUser(ctx context.Context, id string) (*unifi.User, error) {
	var result *unifi.User
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.GetUser(ctx, id)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) CreateUser(ctx context.Context, user *unifi.User) (*unifi.User, error) {
	var result *unifi.User
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateUser(ctx, user)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateUser(ctx context.Context, id string, user *unifi.User) (*unifi.User, error) {
	var result *unifi.User
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateUser(ctx, id, user)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) DeleteUser(ctx context.Context, id string) error {
	return c.withRetry(ctx, func() error {
		return c.client.DeleteUser(ctx, id)
	})
}
