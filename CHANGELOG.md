# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
