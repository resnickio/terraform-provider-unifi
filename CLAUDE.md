# UniFi Terraform Provider

Terraform provider for UniFi network infrastructure management.

## Structure

- `internal/provider/` - Provider implementation
  - `provider.go` - Provider configuration and initialization
  - `client.go` - Auto-relogin client wrapper with retry logic and rate limiting
  - `utils.go` - Pointer helpers, error handling utilities, `stringValueOrNull`
  - `traffic_types.go` - Shared nested types for traffic rules/routes
  - `dynamic_dns_resource.go` - Dynamic DNS configuration resource
  - `dynamic_dns_data_source.go` - Dynamic DNS data source (lookup by ID or hostname)
  - `firewall_group_resource.go` - Address/port group resource
  - `firewall_group_data_source.go` - Firewall group data source (lookup by ID or name)
  - `firewall_policy_resource.go` - Zone-based firewall policy (v2 API)
  - `firewall_policy_data_source.go` - Firewall policy data source (lookup by ID or name)
  - `firewall_rule_resource.go` - Legacy firewall rule resource
  - `firewall_rule_data_source.go` - Firewall rule data source (lookup by ID or name)
  - `firewall_zone_data_source.go` - Firewall zone data source (lookup by ID or name)
  - `firewall_zone_resource.go` - Firewall zone resource (v2 API)
  - `nat_rule_resource.go` - NAT rule resource (v2 API)
  - `nat_rule_data_source.go` - NAT rule data source (lookup by ID or description)
  - `network_data_source.go` - Network data source (lookup by ID or name)
  - `network_resource.go` - Network/VLAN resource
  - `port_forward_resource.go` - Port forwarding resource
  - `port_forward_data_source.go` - Port forward data source (lookup by ID or name)
  - `port_profile_resource.go` - Switch port profile resource
  - `port_profile_data_source.go` - Port profile data source (lookup by ID or name)
  - `radius_profile_resource.go` - RADIUS authentication profile resource
  - `static_dns_resource.go` - Static DNS record resource (v2 API)
  - `static_dns_data_source.go` - Static DNS data source (lookup by ID or key)
  - `static_route_resource.go` - Static route resource
  - `static_route_data_source.go` - Static route data source (lookup by ID or name)
  - `traffic_route_resource.go` - Traffic route (policy-based routing, v2 API)
  - `traffic_route_data_source.go` - Traffic route data source (lookup by ID or name)
  - `traffic_rule_resource.go` - Traffic rule (QoS/blocking, v2 API)
  - `traffic_rule_data_source.go` - Traffic rule data source (lookup by ID or name)
  - `user_resource.go` - User (client device record) resource
  - `user_data_source.go` - User data source (lookup by ID or MAC)
  - `user_group_resource.go` - User group (bandwidth profile) resource
  - `user_group_data_source.go` - User group data source (lookup by ID or name)
  - `wlan_resource.go` - Wireless network (SSID) resource
  - `wlan_data_source.go` - WLAN data source (lookup by ID or name)
  - `testutils_test.go` - Shared test utilities and helpers
  - `sweep_test.go` - Sweeper functions for test resource cleanup
  - `*_test.go` - Acceptance tests for each resource
- `main.go` - Provider entry point

## Dependencies

- [UniFi Go SDK](https://github.com/resnickio/unifi-go-sdk) - API client library
- [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework) - Provider framework (v1.17.0+)
- [terraform-plugin-framework-validators](https://github.com/hashicorp/terraform-plugin-framework-validators) - Input validation for schema attributes
- [terraform-plugin-framework-timeouts](https://github.com/hashicorp/terraform-plugin-framework-timeouts) - Configurable timeout support
- [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing) - Acceptance test framework (v1.14.0+)
- [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) - Documentation generation

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

# Clean up leftover test resources
make sweep

# Generate documentation
make docs

# Format code
make fmt
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

### Resource Compatibility by Controller Type

| Resource | UDM (Network 10.x) | Standalone Network App |
|----------|-------------------|------------------------|
| `unifi_dynamic_dns` | ✅ | ✅ |
| `unifi_firewall_group` | ✅ | ✅ |
| `unifi_firewall_policy` | ✅ | ❌ (500 errors) |
| `unifi_firewall_rule` | ❌ (use zone-based) | ✅ |
| `unifi_firewall_zone` | ✅ | ❌ (500 errors) |
| `unifi_nat_rule` | ✅ | ❌ (v2 API) |
| `unifi_network` | ✅ | ✅ |
| `unifi_port_forward` | ✅ | ✅ |
| `unifi_port_profile` | ✅ | ✅ |
| `unifi_radius_profile` | ✅ | ✅ |
| `unifi_static_dns` | ✅ | ❌ (v2 API) |
| `unifi_static_route` | ✅ | ✅ |
| `unifi_traffic_route` | ✅ | ❌ (v2 API) |
| `unifi_traffic_rule` | ✅ | ❌ (v2 API) |
| `unifi_user` | ✅ | ✅ |
| `unifi_user_group` | ✅ | ✅ |
| `unifi_wlan` | ✅ | ✅ |

**Notes:**
- UDM with Network 10.x uses zone-based firewall (v2 API), legacy rules don't work
- Standalone Network Application may not support zone-based firewall
- Tests auto-skip on unsupported controllers

### Known Limitations

| Resource | Limitation | Reason |
|----------|------------|--------|
| `unifi_firewall_zone` | No `site_id` attribute | UniFi API doesn't return site_id for zones |
| `unifi_firewall_policy` | No `site_id` attribute | UniFi API doesn't return site_id for policies |
| `unifi_wlan` | Import loses passphrase | API never returns passphrase (write-only) |
| `unifi_radius_profile` | Import loses server secrets | API never returns secret field (write-only) |
| `unifi_dynamic_dns` | Import loses password | API never returns password (write-only) |

Site limitations: Resources still work correctly - the site is determined by the provider's `site` configuration.

Write-only limitations: After import, users must re-apply configuration to set these values, or use `terraform state` commands to manually populate them.

## Provider Architecture

### Auto-Relogin Client

The provider wraps the SDK client with automatic re-authentication:

```go
type AutoLoginClient struct {
    client       unifi.NetworkManager
    config       unifi.NetworkClientConfig
    mu           sync.Mutex
    lastAuthTime time.Time
    authSem      chan struct{}
}
```

Uses channel-based semaphore for context-aware rate limiting and concurrent re-authentication handling.

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

### Write-Only Fields

Some API fields (passwords, secrets, passphrases) are write-only - the API accepts them but never returns them. These require special handling to prevent state drift:

1. **Schema**: Mark as `Sensitive: true` and add `UseStateForUnknown()` plan modifier
2. **Create/Update**: Save value from plan before calling `sdkToState()`, restore after
3. **Read**: Pass `priorState` to `sdkToState()` to preserve value
4. **sdkToState()**: Check if API returned value, otherwise preserve from prior state

Resources using this pattern:
- `unifi_wlan` - `passphrase` field
- `unifi_radius_profile` - `secret` field in auth/acct server blocks
- `unifi_dynamic_dns` - `password` field

### Data Source Pattern

Each data source follows this pattern:

1. **Model struct** - Same as resource model but all fields are Computed
2. **Schema** - Uses `AtLeastOneOf` validator for id/lookup-field (name, key, hostname, description)
3. **Read method** - Fetches by ID directly, or lists all and filters by lookup field
4. **sdkToState()** - Reuses resource's conversion function where possible

### Nested Attributes vs Blocks

**Always use nested attributes, not blocks, for new schema implementations.**

Per [HashiCorp's official guidance](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks/list-nested):
> "Use nested attribute types instead of block types for new schema implementations. Block support is mainly for migrating legacy SDK-based providers."

**Use in schema:**
```go
// Correct: ListNestedAttribute / SingleNestedAttribute
"servers": schema.ListNestedAttribute{
    Optional: true,
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{...},
    },
}

// Wrong: ListNestedBlock / SingleNestedBlock (legacy pattern)
"servers": schema.ListNestedBlock{
    NestedObject: schema.NestedBlockObject{...},
}
```

**HCL syntax in tests:**
```hcl
# Correct: attribute syntax with = and brackets
servers = [{
  ip   = "10.0.0.1"
  port = 1812
}]

schedule = {
  mode = "ALWAYS"
}

# Wrong: block syntax (no =, no brackets)
servers {
  ip   = "10.0.0.1"
}
```

## Preferences

- **Commits**: Do not include Claude Code citations or co-author tags
- **Code style**: Minimal comments, no inline comments unless truly necessary
- **Over-engineering**: Avoid. Don't add abstractions, helpers, or features beyond what's requested
- **Resource naming**: Test resources use `tf-acc-test-` prefix for easy identification
- **VLAN IDs**: Use 3900+ range in tests to avoid production conflicts
- **Rule indices**: Use 2000+ range in tests (must start with 2 or 4 per API validation)
- **Context7 MCP**: When generating code that uses external libraries, or when needing up-to-date API documentation, configuration examples, or setup steps for any library/framework, automatically use Context7 MCP tools (`resolve-library-id` then `get-library-docs`) to fetch current documentation. Do not rely solely on training data for library APIs.
- **Playwright MCP**: Use Playwright MCP tools for browser automation tasks: testing web UIs, scraping dynamic content, filling forms, taking screenshots, or interacting with web applications. Prefer `browser_snapshot` over screenshots for actionable page state. Use `browser_fill_form` for multiple fields, `browser_click`/`browser_type` for interactions, and `browser_evaluate` for custom JavaScript. Always call `browser_close` when finished.

## Testing Conventions

- **Acceptance tests**: Comprehensive coverage for all attributes including:
  - All attribute combinations
  - Default value verification
  - Update/modification scenarios
  - Import state verification
- **Test helpers**: Use shared helpers from `testutils_test.go` (`testAccPreCheck`, `testAccCheckResourceDestroy`, etc.)
- **Controller compatibility**: Use `testAccPreCheckZoneBasedFirewall` or similar prechecks for controller-specific features
- **Resource cleanup**: Implement sweeper functions in `sweep_test.go` for test resource cleanup
- **Naming**: Test resources use `tf-acc-test-` prefix, test functions use `TestAcc{Resource}_{scenario}` pattern
- **Config builders**: Use helper functions to build test configs with consistent patterns

## Post-Implementation Summaries

After completing a planned task, provide a concise summary including:
- **Files Created**: New files with brief descriptions
- **Files Modified**: Existing files and what changed
- **Key Details**: Coverage, scope, or other relevant metrics

## Status

**Implemented Resources:**
- `unifi_dynamic_dns` - Dynamic DNS configuration
- `unifi_firewall_group` - Address and port groups
- `unifi_firewall_policy` - Zone-based firewall (v2 API)
- `unifi_firewall_rule` - Legacy firewall rules
- `unifi_firewall_zone` - Firewall zones (v2 API)
- `unifi_nat_rule` - NAT rules (v2 API)
- `unifi_network` - VLAN networks with DHCP
- `unifi_port_forward` - Port forwarding rules
- `unifi_port_profile` - Switch port profiles (VLAN, PoE, 802.1X, storm control)
- `unifi_radius_profile` - RADIUS authentication profiles for 802.1X
- `unifi_static_dns` - Static DNS records (v2 API)
- `unifi_static_route` - Static routing
- `unifi_traffic_route` - Traffic routes/policy-based routing (v2 API)
- `unifi_traffic_rule` - Traffic rules for QoS/blocking (v2 API)
- `unifi_user` - Client device records (DHCP reservations, fixed IPs, device names, blocking)
- `unifi_user_group` - Bandwidth/QoS groups
- `unifi_wlan` - Wireless networks (SSID configuration)

**Implemented Data Sources:**
- `unifi_dynamic_dns` - Look up dynamic DNS configuration by ID or hostname
- `unifi_firewall_group` - Look up firewall group (address/port) by ID or name
- `unifi_firewall_policy` - Look up firewall policy by ID or name
- `unifi_firewall_rule` - Look up firewall rule by ID or name
- `unifi_firewall_zone` - Look up firewall zone by ID or name
- `unifi_nat_rule` - Look up NAT rule by ID or description
- `unifi_network` - Look up network by ID or name
- `unifi_port_forward` - Look up port forward by ID or name
- `unifi_port_profile` - Look up port profile by ID or name
- `unifi_static_dns` - Look up static DNS record by ID or key (hostname)
- `unifi_static_route` - Look up static route by ID or name
- `unifi_traffic_route` - Look up traffic route by ID or name
- `unifi_traffic_rule` - Look up traffic rule by ID or name
- `unifi_user` - Look up user (client device record) by ID or MAC
- `unifi_user_group` - Look up user group by ID or name
- `unifi_wlan` - Look up WLAN by ID or name

## Versioning and Releases

This provider follows [Semantic Versioning](https://semver.org/):
- **v0.x.x** - Development phase, breaking changes allowed between minor versions
- **v1.0.0+** - Stable API, breaking changes only in major versions

### Current Status

- Provider is in **v0.x** development phase
- Breaking changes should increment minor version (v0.1.0 → v0.2.0)
- Bug fixes increment patch version (v0.1.0 → v0.1.1)

### Branching Strategy

- **main** - Always deployable, protected branch
- **feature/*** - Short-lived feature branches, merge via PR
- Tags mark releases (e.g., `v0.1.0`, `v0.2.0`)

### Release Process

1. Ensure all tests pass: `make test && make testacc`
2. Ensure build succeeds: `make build`
3. Regenerate docs: `make docs`
4. Create annotated tag: `git tag -a vX.Y.Z -m "Release description"`
5. Push tag: `git push origin vX.Y.Z`

### Registry Publishing

For Terraform Registry publication:
- Ensure all resources have import documentation
- Ensure all resources have example configurations in `examples/`
- Verify generated docs in `docs/` are complete
- Follow HashiCorp's [publishing requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing)

## Related Projects

- [UniFi Go SDK](https://github.com/resnickio/unifi-go-sdk) - Sister project, the underlying API client

Reference for patterns and lessons learned (not for copying code):
- **paultyng/terraform-provider-unifi**: Original community provider (abandoned). Uses older SDK patterns.
- **ubiquiti-community fork**: Maintenance-only fork. Validates need for fresh implementation.
- **filipowm/unifi**: Has known data loss bugs. Demonstrates importance of proper state handling.
