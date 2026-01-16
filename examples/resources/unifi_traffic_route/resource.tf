# Basic traffic route
resource "unifi_traffic_route" "basic" {
  name        = "VPN-Route"
  description = "Route traffic through VPN"
  network_id  = unifi_network.vpn.id
}

# Domain-based routing
resource "unifi_traffic_route" "domain_route" {
  name            = "Work-Domains"
  matching_target = "DOMAIN"
  network_id      = unifi_network.work_vpn.id

  domains {
    domain      = "*.work.example.com"
    description = "Work domain"
  }

  domains {
    domain = "corporate.example.com"
  }

  kill_switch = true
}

# IP-based routing
resource "unifi_traffic_route" "ip_route" {
  name            = "Datacenter-Route"
  matching_target = "IP"
  network_id      = unifi_network.datacenter.id
  ip_addresses    = ["10.0.0.0/8", "172.16.0.0/12"]
}

# Region-based routing (geo-routing)
resource "unifi_traffic_route" "region_route" {
  name            = "US-Traffic"
  matching_target = "REGION"
  network_id      = unifi_network.us_vpn.id
  regions         = ["US", "CA"]
}

# Route with fallback enabled
resource "unifi_traffic_route" "fallback_route" {
  name       = "Backup-Route"
  network_id = unifi_network.backup.id
  fallback   = true
}

# Disabled route
resource "unifi_traffic_route" "disabled" {
  name    = "Maintenance-Route"
  enabled = false
}
