# UniFi Terraform Provider

[![CI](https://github.com/resnickio/terraform-provider-unifi/actions/workflows/ci.yml/badge.svg)](https://github.com/resnickio/terraform-provider-unifi/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/resnickio/terraform-provider-unifi)](https://goreportcard.com/report/github.com/resnickio/terraform-provider-unifi)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Terraform provider for managing UniFi network infrastructure using the [UniFi Go SDK](https://github.com/resnickio/unifi-go-sdk).

## Purpose

This provider enables declarative infrastructure-as-code management of UniFi network configurations. It connects directly to a UniFi controller (Dream Machine, Cloud Key, or standalone controller) to manage networks, firewall rules, port forwards, and more.

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for building from source)
- UniFi Controller with local access

## Installation

### From Source

```bash
git clone https://github.com/resnickio/terraform-provider-unifi.git
cd terraform-provider-unifi
make install
```

## Provider Configuration

### Using API Key (Recommended)

```hcl
provider "unifi" {
  base_url = "https://192.168.1.1"
  api_key  = "your-api-key"
  site     = "default"    # optional, defaults to "default"
  insecure = true         # optional, for self-signed certs
}
```

### Using Username/Password

```hcl
provider "unifi" {
  base_url = "https://192.168.1.1"
  username = "admin"
  password = "password"
  site     = "default"    # optional, defaults to "default"
  insecure = true         # optional, for self-signed certs
}
```

### Environment Variables

All configuration can be set via environment variables:

| Variable | Description |
|----------|-------------|
| `UNIFI_BASE_URL` | Controller URL (e.g., `https://192.168.1.1`) |
| `UNIFI_API_KEY` | API key for authentication (recommended) |
| `UNIFI_USERNAME` | Admin username (alternative to API key) |
| `UNIFI_PASSWORD` | Admin password (alternative to API key) |
| `UNIFI_SITE` | Site name (default: `default`) |
| `UNIFI_INSECURE` | Skip TLS verification (`true`/`false`) |

API key authentication is recommended and takes priority over username/password when both are provided.

## Resources

### unifi_network

Manages VLAN networks with DHCP configuration.

```hcl
resource "unifi_network" "iot" {
  name         = "IoT Network"
  purpose      = "corporate"
  vlan_id      = 100
  subnet       = "10.0.100.0/24"
  dhcp_enabled = true
  dhcp_start   = "10.0.100.10"
  dhcp_stop    = "10.0.100.254"
  dhcp_lease   = 86400
  dhcp_dns     = ["8.8.8.8", "8.8.4.4"]
  domain_name  = "iot.local"
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Network name |
| `purpose` | string | yes | `corporate`, `guest`, `wan`, or `vlan-only` |
| `vlan_id` | number | no | VLAN ID (1-4095) |
| `subnet` | string | no | CIDR notation (e.g., `10.0.100.0/24`) |
| `dhcp_enabled` | bool | no | Enable DHCP server (default: `true`) |
| `dhcp_start` | string | no | DHCP range start IP |
| `dhcp_stop` | string | no | DHCP range end IP |
| `dhcp_lease` | number | no | Lease time in seconds (default: `86400`) |
| `dhcp_dns` | list | no | DNS servers for DHCP clients |
| `domain_name` | string | no | Domain name for the network |
| `igmp_snooping` | bool | no | Enable IGMP snooping (default: `false`) |
| `enabled` | bool | no | Enable the network (default: `true`) |

### unifi_firewall_group

Manages address and port groups for firewall rules.

```hcl
resource "unifi_firewall_group" "blocked_ips" {
  name       = "Blocked IPs"
  group_type = "address-group"
  members    = ["1.2.3.4", "5.6.7.8", "10.0.0.0/24"]
}

resource "unifi_firewall_group" "web_ports" {
  name       = "Web Ports"
  group_type = "port-group"
  members    = ["80", "443", "8080-8090"]
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Group name |
| `group_type` | string | yes | `address-group` or `port-group` |
| `members` | list | yes | IP addresses/CIDRs or ports/ranges |

### unifi_firewall_rule

Manages legacy firewall rules.

```hcl
resource "unifi_firewall_rule" "block_iot_to_lan" {
  name       = "Block IoT to LAN"
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = 4000
  enabled    = true
  protocol   = "all"

  src_firewall_group_ids = [unifi_firewall_group.iot_devices.id]
  dst_firewall_group_ids = [unifi_firewall_group.lan_devices.id]
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Rule name |
| `ruleset` | string | yes | `LAN_IN`, `LAN_OUT`, `WAN_IN`, `WAN_OUT`, `WAN_LOCAL`, etc. |
| `action` | string | yes | `accept`, `drop`, or `reject` |
| `rule_index` | number | yes | Rule priority (lower = higher priority) |
| `enabled` | bool | no | Enable the rule (default: `true`) |
| `protocol` | string | no | `all`, `tcp`, `udp`, `tcp_udp`, `icmp` (default: `all`) |
| `src_address` | string | no | Source IP or CIDR |
| `src_firewall_group_ids` | list | no | Source firewall group IDs |
| `dst_address` | string | no | Destination IP or CIDR |
| `dst_firewall_group_ids` | list | no | Destination firewall group IDs |
| `dst_port` | string | no | Destination port or range |
| `logging` | bool | no | Log matching traffic (default: `false`) |

### unifi_port_forward

Manages port forwarding rules.

```hcl
resource "unifi_port_forward" "web_server" {
  name     = "Web Server"
  protocol = "tcp"
  dst_port = "443"
  fwd_ip   = "10.0.0.50"
  fwd_port = "443"
  enabled  = true
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Rule name |
| `protocol` | string | yes | `tcp`, `udp`, or `tcp_udp` |
| `dst_port` | string | yes | External port or range |
| `fwd_ip` | string | yes | Internal destination IP |
| `fwd_port` | string | yes | Internal destination port |
| `enabled` | bool | no | Enable the rule (default: `true`) |
| `src` | string | no | Source IP/CIDR restriction |
| `pfwd_interface` | string | no | WAN interface: `wan`, `wan2`, `both` (default: `wan`) |
| `log` | bool | no | Log forwarded traffic (default: `false`) |

## Import

All resources support import by ID:

```bash
terraform import unifi_network.example 60a1b2c3d4e5f6a7b8c9d0e1
terraform import unifi_firewall_group.example 60a1b2c3d4e5f6a7b8c9d0e2
terraform import unifi_firewall_rule.example 60a1b2c3d4e5f6a7b8c9d0e3
terraform import unifi_port_forward.example 60a1b2c3d4e5f6a7b8c9d0e4
```

## Development

### Build

```bash
make build
```

### Test

```bash
# Unit tests
make test

# Acceptance tests (requires UniFi controller)
cp .env.example .env
# Edit .env with your controller credentials
make testacc
```

### Install Locally

```bash
make install
```

## Data Sources

All data sources support lookup by ID or a resource-specific field (usually `name`).

```hcl
# Look up by name
data "unifi_network" "lan" {
  name = "Default"
}

# Or look up by ID
data "unifi_network" "by_id" {
  id = "60a1b2c3d4e5f6a7b8c9d0e1"
}
```

| Data Source | Lookup Fields | Description |
|-------------|---------------|-------------|
| `unifi_network` | `id`, `name` | Network/VLAN configuration |
| `unifi_firewall_group` | `id`, `name` | Address and port groups |
| `unifi_firewall_zone` | `id`, `name` | Firewall zones (v2 API) |
| `unifi_firewall_rule` | `id`, `name` | Legacy firewall rules |
| `unifi_firewall_policy` | `id`, `name` | Zone-based policies (v2 API) |
| `unifi_port_forward` | `id`, `name` | Port forwarding rules |
| `unifi_port_profile` | `id`, `name` | Switch port profiles |
| `unifi_static_route` | `id`, `name` | Static routes |
| `unifi_static_dns` | `id`, `key` | Static DNS records (v2 API) |
| `unifi_dynamic_dns` | `id`, `hostname` | Dynamic DNS configuration |
| `unifi_nat_rule` | `id`, `description` | NAT rules (v2 API) |
| `unifi_traffic_route` | `id`, `description` | Traffic routes (v2 API) |
| `unifi_traffic_rule` | `id`, `description` | Traffic rules (v2 API) |
| `unifi_user` | `id`, `mac` | Client device records |
| `unifi_user_group` | `id`, `name` | User/bandwidth groups |
| `unifi_wlan` | `id`, `name` | Wireless networks |

See `docs/data-sources/` for detailed attribute documentation.

## Status

### Resources (17)

| Resource | Description | API |
|----------|-------------|-----|
| `unifi_network` | VLAN networks with DHCP | v1 |
| `unifi_firewall_group` | Address and port groups | v1 |
| `unifi_firewall_zone` | Firewall zones | v2 |
| `unifi_firewall_policy` | Zone-based firewall rules | v2 |
| `unifi_firewall_rule` | Legacy firewall rules | v1 |
| `unifi_port_forward` | Port forwarding | v1 |
| `unifi_port_profile` | Switch port profiles | v1 |
| `unifi_static_route` | Static routing | v1 |
| `unifi_static_dns` | Static DNS records | v2 |
| `unifi_dynamic_dns` | Dynamic DNS | v1 |
| `unifi_nat_rule` | NAT rules | v2 |
| `unifi_traffic_route` | Policy-based routing | v2 |
| `unifi_traffic_rule` | QoS and traffic blocking | v2 |
| `unifi_radius_profile` | RADIUS authentication | v1 |
| `unifi_user` | Client device records | v1 |
| `unifi_user_group` | Bandwidth/QoS groups | v1 |
| `unifi_wlan` | Wireless networks (SSIDs) | v1 |

### Data Sources (18)

All resources above have corresponding data sources, plus the `unifi_device` data source for device lookup.

### Controller Compatibility

| API Version | UDM/Cloud Key | Standalone Controller |
|-------------|---------------|----------------------|
| v1 resources | ✅ | ✅ |
| v2 resources | ✅ | ❌ |

See `docs/` for detailed documentation on each resource and data source.

## Related Projects

- [UniFi Go SDK](https://github.com/resnickio/unifi-go-sdk) - The underlying SDK used by this provider

## License

MIT

## Development

This provider was developed with AI assistance. If you encounter bugs or have feature requests, please [open an issue](https://github.com/resnickio/terraform-provider-unifi/issues).
