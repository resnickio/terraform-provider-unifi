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
}

// NewAutoLoginClient creates a new auto-login wrapper around the SDK client.
func NewAutoLoginClient(client unifi.NetworkManager, config unifi.NetworkClientConfig) *AutoLoginClient {
	return &AutoLoginClient{
		client: client,
		config: config,
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

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: if we already re-authenticated after this request started,
	// just retry without re-authenticating again
	if c.lastAuthTime.After(failedAt) {
		return fn()
	}

	// Rate limit: don't re-auth more than once per minAuthInterval
	if timeSinceLastAuth := time.Since(c.lastAuthTime); timeSinceLastAuth < minAuthInterval {
		time.Sleep(minAuthInterval - timeSinceLastAuth)
	}

	// Re-authenticate
	if loginErr := c.client.Login(ctx); loginErr != nil {
		return fmt.Errorf("re-authentication failed: %w", loginErr)
	}
	c.lastAuthTime = time.Now()

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

func (c *AutoLoginClient) CreateFirewallZone(ctx context.Context, zone *unifi.FirewallZone) (*unifi.FirewallZone, error) {
	var result *unifi.FirewallZone
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.CreateFirewallZone(ctx, zone)
		return err
	})
	return result, err
}

func (c *AutoLoginClient) UpdateFirewallZone(ctx context.Context, id string, zone *unifi.FirewallZone) (*unifi.FirewallZone, error) {
	var result *unifi.FirewallZone
	err := c.withRetry(ctx, func() error {
		var err error
		result, err = c.client.UpdateFirewallZone(ctx, id, zone)
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
