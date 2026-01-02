# UniFi Terraform Provider

Terraform provider for UniFi network infrastructure management.

## Structure

- `internal/provider/` - Provider implementation
  - `provider.go` - Provider configuration and initialization
  - `client.go` - Auto-relogin client wrapper with retry logic and rate limiting
  - `utils.go` - Pointer helpers, error handling utilities, `stringValueOrNull`
  - `network_resource.go` - Network/VLAN resource
  - `network_data_source.go` - Network data source (lookup by ID or name)
  - `firewall_group_resource.go` - Address/port group resource
  - `firewall_rule_resource.go` - Legacy firewall rule resource
  - `firewall_policy_resource.go` - Zone-based firewall policy (v2 API)
  - `firewall_zone_resource.go` - Firewall zone resource (v2 API)
  - `port_forward_resource.go` - Port forwarding resource
  - `static_route_resource.go` - Static route resource
  - `user_group_resource.go` - User group (bandwidth profile) resource
  - `wlan_resource.go` - Wireless network (SSID) resource
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
| `unifi_network` | ✅ | ✅ |
| `unifi_firewall_group` | ✅ | ✅ |
| `unifi_firewall_rule` | ❌ (use zone-based) | ✅ |
| `unifi_firewall_policy` | ✅ | ❌ (500 errors) |
| `unifi_firewall_zone` | ✅ | ❌ (500 errors) |
| `unifi_port_forward` | ✅ | ✅ |
| `unifi_static_route` | ✅ | ✅ |
| `unifi_user_group` | ✅ | ✅ |
| `unifi_wlan` | ✅ | ✅ |

**Notes:**
- UDM with Network 10.x uses zone-based firewall (v2 API), legacy rules don't work
- Standalone Network Application may not support zone-based firewall
- Tests auto-skip on unsupported controllers

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
- `unifi_network` - VLAN networks with DHCP
- `unifi_firewall_group` - Address and port groups
- `unifi_firewall_rule` - Legacy firewall rules
- `unifi_firewall_policy` - Zone-based firewall (v2 API)
- `unifi_firewall_zone` - Firewall zones (v2 API)
- `unifi_port_forward` - Port forwarding rules
- `unifi_static_route` - Static routing
- `unifi_user_group` - Bandwidth/QoS groups
- `unifi_wlan` - Wireless networks (SSID configuration)

**Implemented Data Sources:**
- `unifi_network` - Look up network by ID or name

**Planned Resources:**
- `unifi_radius_profile` - RADIUS authentication profiles
- `unifi_dynamic_dns` - Dynamic DNS configuration
- `unifi_port_profile` - Switch port profiles

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
