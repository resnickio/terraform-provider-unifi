# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.0] - 2026-05-02

### Fixed

- `unifi_firewall_policy` — **silent data loss when `matching_target` defaults to `ANY`**. The schema's static `"ANY"` default let users send `matching_target=ANY` alongside `ips=[...]`; the UniFi controller silently discarded `ips` and widened the policy to allow ANY destination in the zone. The `terraform plan` diff looked cosmetic (`matching_target: "IP" -> "ANY"`) so a careful operator could not tell the policy was about to widen. **Operators with imported policies should re-plan after upgrading; an unexpected `matching_target` change in the diff means the prior controller-side state had already been mutated.** ([#2](https://github.com/resnickio/terraform-provider-unifi/pull/2))
- `unifi_firewall_policy` — controller now requires `matching_target_type` (`SPECIFIC` for inline values, `OBJECT` for group references) on every IP/NETWORK match. The provider was not sending this field, causing `api.err.MissingFirewallDestinationMatchingTargetType` from v9 controllers. The provider now derives it transparently from `matching_target`. ([#2](https://github.com/resnickio/terraform-provider-unifi/pull/2))
- `unifi_network` — `mdns_enabled` and `upnp_lan_enabled` were marked `Optional` only with no default, so omitting them from config caused `Provider produced inconsistent result after apply: was null, but now cty.False` whenever the controller returned the field. Both attributes are now `Optional + Computed` and resolve to whatever the controller persists. ([#3](https://github.com/resnickio/terraform-provider-unifi/pull/3))

### Changed

- `unifi_firewall_policy` — `source.matching_target` and `destination.matching_target` are now auto-derived from sibling fields when the user leaves them unset: `ips` non-empty → `"IP"`, `network_id` non-empty → `"NETWORK"`, otherwise `"ANY"`. Users who previously specified `matching_target` explicitly are unaffected; users who left it unset will see plans correctly resolve to `IP`/`NETWORK` instead of the broken `ANY` default.
- `unifi_firewall_policy` — `matching_target` now validates against the enum at plan time via `stringvalidator.OneOf(...)`. Typos that previously reached the controller now fail in `terraform plan` with a clear error.
- `unifi_firewall_policy` — explicit `matching_target = "ANY"` alongside `ips` or `network_id` is now rejected at plan time as a config error, since UniFi silently strips those fields under `matching_target=ANY`.

### Known Issues

- `unifi_firewall_policy` — matching targets `DOMAIN`, `REGION`, `PORT_GROUP`, `ADDRESS_GROUP` are not usable. The Go SDK's `PolicyEndpoint` type lacks fields to carry the corresponding match data (domains, regions, group IDs); the provider can set the marker but not the values. Manage these policies via the UniFi UI for now.
- `unifi_firewall_policy` — changing `action` from `ALLOW` to `BLOCK` in a single apply can fail on UniFi v9 with `api.err.FirewallPolicyCreateRespondTrafficPolicyNotAllowed`. Workaround: destroy and recreate the policy. Pre-existing; not introduced by 0.8.0.
- `unifi_network` — site-wide `unifi_setting_usg.mdns_enabled` gates whether per-network `mdns_enabled = true` takes effect. The controller silently strips per-network `true` if the site-level toggle is off, with no error.

## [0.4.0] - 2026-02-14

### Added

#### New Resources
- `unifi_user` - Client device records with DHCP reservations, fixed IPs, device naming, local DNS records, blocking, and user group assignments

#### New Data Sources
- `unifi_user` - Look up user by ID or MAC address

### Changed
- Updated unifi-go-sdk from v0.5.0 to v0.6.0 (adds User CRUD support)
- Provider now registers 17 resources and 18 data sources

## [0.2.1] - 2025-01-16

### Fixed

- `unifi_port_profile` - Corrected `tagged_vlan_mgmt` valid values to "auto", "block_all", "custom" (was incorrectly "all", "block", "custom")
- Test fixes for dynamic DNS (use "custom" service which is universally supported)
- Test fixes for traffic rules (added required `matching_target` and `target_devices` fields)
- Test fixes for WLAN data source (removed reference to non-existent `unifi_ap_group` data source)
- Test fixes for firewall rules (added precheck to skip on zone-based firewall controllers)

### Known Issues

- `unifi_traffic_rule` - Provider has state consistency issues with default values for `bandwidth_limit` and `schedule`
- `unifi_traffic_route` - Similar state consistency issues
- `unifi_static_dns`, `unifi_nat_rule`, `unifi_radius_profile` - Schema issues with nested blocks vs attributes

## [0.2.0] - 2025-01-16

### Added

#### New Resources
- `unifi_static_dns` - Static DNS records (v2 API)
- `unifi_dynamic_dns` - Dynamic DNS configuration
- `unifi_nat_rule` - NAT rules (v2 API)
- `unifi_traffic_rule` - Traffic rules for QoS and blocking (v2 API)
- `unifi_traffic_route` - Traffic routes for policy-based routing (v2 API)
- `unifi_radius_profile` - RADIUS authentication profiles
- `unifi_port_profile` - Switch port profiles with VLAN, PoE, 802.1X, and storm control

#### New Data Sources
- `unifi_static_dns` - Look up static DNS by ID or key
- `unifi_dynamic_dns` - Look up dynamic DNS by ID or hostname
- `unifi_nat_rule` - Look up NAT rule by ID or description
- `unifi_traffic_rule` - Look up traffic rule by ID or description
- `unifi_traffic_route` - Look up traffic route by ID or description
- `unifi_port_profile` - Look up port profile by ID or name
- `unifi_port_forward` - Look up port forward by ID or name
- `unifi_firewall_policy` - Look up firewall policy by ID or name
- `unifi_firewall_rule` - Look up firewall rule by ID or name
- `unifi_static_route` - Look up static route by ID or name
- `unifi_user_group` - Look up user group by ID or name
- `unifi_wlan` - Look up WLAN by ID or name

#### Documentation
- Added example configurations for all data sources
- Expanded CLAUDE.md with write-only fields pattern and data source pattern
- Updated README.md with complete resource and data source listing

### Fixed
- WLAN passphrase now correctly preserved during Update operations (previously caused state drift)

### Changed
- Provider now registers 16 resources and 15 data sources (up from 9 resources and 3 data sources)

## [0.1.1] - 2025-01-14

### Added
- `unifi_firewall_zone` data source
- `unifi_firewall_group` data source

### Fixed
- Documentation improvements

## [0.1.0] - 2025-01-13

### Added

#### Resources
- `unifi_network` - VLAN networks with DHCP configuration
- `unifi_firewall_group` - Address and port groups
- `unifi_firewall_rule` - Legacy firewall rules
- `unifi_firewall_policy` - Zone-based firewall policies (v2 API)
- `unifi_firewall_zone` - Firewall zones (v2 API)
- `unifi_port_forward` - Port forwarding rules
- `unifi_static_route` - Static routing
- `unifi_user_group` - User/bandwidth groups
- `unifi_wlan` - Wireless networks (SSIDs)

#### Data Sources
- `unifi_network` - Look up network by ID or name

#### Infrastructure
- Auto-relogin client with retry logic and rate limiting
- Support for both API key and username/password authentication
- Configurable timeouts for all CRUD operations
- Comprehensive acceptance test suite with sweepers
