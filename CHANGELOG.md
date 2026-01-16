# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
