# UniFi Terraform Provider

Terraform provider for UniFi network infrastructure management.

## Structure

- `internal/provider/` - Provider implementation
  - `provider.go` - Provider configuration and initialization
  - `client.go` - Auto-relogin client wrapper with retry logic
  - `utils.go` - Pointer helpers, error handling utilities
  - `network_resource.go` - Network/VLAN resource
  - `firewall_group_resource.go` - Address/port group resource
  - `firewall_rule_resource.go` - Legacy firewall rule resource
  - `firewall_policy_resource.go` - Zone-based firewall policy (v2 API)
  - `firewall_zone_resource.go` - Firewall zone resource (v2 API)
  - `port_forward_resource.go` - Port forwarding resource
  - `static_route_resource.go` - Static route resource
  - `user_group_resource.go` - User group (bandwidth profile) resource
  - `wlan_resource.go` - Wireless network (SSID) resource
  - `*_test.go` - Acceptance tests for each resource
- `main.go` - Provider entry point

## Dependencies

- [UniFi Go SDK](https://github.com/resnickio/unifi-go-sdk) - API client library
- [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) - Provider framework (v1.17.0+)
- [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing) - Acceptance test framework (v1.14.0+)

## Build & Test

```bash
# Build
make build

# Unit tests
make test

# Acceptance tests (requires .env with credentials)
make testacc

# Run specific test
make testacc-run TEST=TestAccNetworkResource_basic
```

## Environment Variables

For acceptance tests, create `.env` from `.env.example`:

```
UNIFI_BASE_URL=https://192.168.1.1
UNIFI_API_KEY=your-api-key        # Recommended - API key authentication
UNIFI_USERNAME=admin               # Alternative - username/password auth
UNIFI_PASSWORD=your-password       # Alternative - username/password auth
UNIFI_INSECURE=true
```

API key authentication is recommended and takes priority over username/password when both are provided.

## UniFi Controller Compatibility

Two controller types with different API paths:

| Platform | API Path | Notes |
|----------|----------|-------|
| UniFi OS (UDM, UDM Pro, Cloud Key Gen2+) | `/proxy/network/api/` | Unified login |
| Legacy Controller (USG, standalone) | `/api/` | Direct controller |

The SDK handles path differences. Both use session-based authentication.

## Provider Architecture

### Auto-Relogin Client

The provider wraps the SDK client with automatic re-authentication:

```go
type AutoLoginClient struct {
    client       unifi.NetworkManager
    config       unifi.NetworkClientConfig
    mu           sync.Mutex
    lastAuthTime time.Time
}
```

Uses double-checked locking pattern to handle concurrent re-authentication safely.

### Resource Pattern

Each resource follows this pattern:

1. **Model struct** - Terraform state representation with `tfsdk` tags
2. **planToSDK()** - Convert Terraform plan to SDK struct
3. **sdkToState()** - Convert SDK response to Terraform state (returns `diag.Diagnostics`)
4. **CRUD methods** - Create, Read, Update, Delete implementations
5. **ImportState** - Passthrough ID import

### Error Handling

- Use `handleSDKError()` for consistent error messages
- Use `isNotFoundError()` to detect 404s for graceful deletion/drift handling
- `planToSDK()` accepts `*diag.Diagnostics` for propagating conversion errors
- `sdkToState()` returns `diag.Diagnostics` for propagating conversion errors

## Preferences

- **Commits**: Do not include Claude Code citations or co-author tags
- **Code style**: Minimal comments, no inline comments unless truly necessary
- **Over-engineering**: Avoid. Don't add abstractions, helpers, or features beyond what's requested
- **Testing**: Comprehensive acceptance tests for all attributes. Tests should cover:
  - All attribute combinations
  - Default value verification
  - Update/modification scenarios
  - Import state verification
- **Resource naming**: Test resources use `tf-acc-test-` prefix for easy identification
- **VLAN IDs**: Use 3900+ range in tests to avoid production conflicts
- **Rule indices**: Use 4000+ range in tests to avoid production conflicts
- **Context7 MCP**: When generating code that uses external libraries, use Context7 MCP tools to fetch current documentation
- **Playwright MCP**: Use for browser automation tasks when needed

## Status

**Implemented Resources:**
- `unifi_network` - VLAN networks with DHCP
- `unifi_firewall_group` - Address and port groups
- `unifi_firewall_rule` - Legacy firewall rules
- `unifi_firewall_policy` - Zone-based firewall (v2 API)
- `unifi_firewall_zone` - Firewall zones (v2 API)
- `unifi_port_forward` - Port forwarding rules
- `unifi_static_route` - Static routing
- `unifi_user_group` - Bandwidth/QoS groups
- `unifi_wlan` - Wireless networks (SSID configuration)

**Planned Resources:**
- `unifi_radius_profile` - RADIUS authentication profiles
- `unifi_dynamic_dns` - Dynamic DNS configuration
- `unifi_port_profile` - Switch port profiles

## Related Projects

- [UniFi Go SDK](https://github.com/resnickio/unifi-go-sdk) - Sister project, the underlying API client
