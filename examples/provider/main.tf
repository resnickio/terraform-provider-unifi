terraform {
  required_providers {
    unifi = {
      source  = "resnickio/unifi"
      version = "0.1.0"
    }
  }
}

provider "unifi" {
  base_url = "https://192.168.1.1"
  api_key  = "your-api-key"  # Recommended: use API key authentication
  site     = "default"
  insecure = true            # Set to false in production with valid TLS certificate

  # Alternative: username/password authentication (not recommended)
  # username = "admin"
  # password = "password"
}

# Example: Create a VLAN network
resource "unifi_network" "iot_network" {
  name         = "IoT Network"
  purpose      = "corporate"
  vlan_id      = 100
  subnet       = "10.0.100.0/24"
  dhcp_enabled = true
  dhcp_start   = "10.0.100.10"
  dhcp_stop    = "10.0.100.254"
  dhcp_dns     = ["8.8.8.8", "8.8.4.4"]
  domain_name  = "iot.local"
}

# Example: Create a firewall group for IoT devices
resource "unifi_firewall_group" "iot_devices" {
  name       = "IoT Devices"
  group_type = "address-group"
  members = [
    "10.0.100.10",
    "10.0.100.11",
    "10.0.100.12",
  ]
}

# Example: Create a firewall group for web ports
resource "unifi_firewall_group" "web_ports" {
  name       = "Web Ports"
  group_type = "port-group"
  members    = ["80", "443", "8080"]
}

# Example: Create a firewall rule to block IoT from accessing LAN
resource "unifi_firewall_rule" "block_iot_to_lan" {
  name       = "Block IoT to LAN"
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = 2000
  enabled    = true

  src_firewall_group_ids = [unifi_firewall_group.iot_devices.id]

  logging = true
}

# Example: Create a port forward for a web server
resource "unifi_port_forward" "web_server" {
  name          = "Web Server"
  enabled       = true
  protocol      = "tcp"
  dst_port      = "443"
  fwd_port      = "443"
  fwd_ip        = "10.0.100.50"
  pfwd_interface = "wan"
  log           = false
}
