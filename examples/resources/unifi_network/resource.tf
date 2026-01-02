# Create a VLAN network with DHCP
resource "unifi_network" "iot" {
  name         = "IoT Network"
  purpose      = "corporate"
  vlan_id      = 100
  subnet       = "10.0.100.1/24"
  dhcp_enabled = true
  dhcp_start   = "10.0.100.10"
  dhcp_stop    = "10.0.100.254"
  dhcp_dns     = ["8.8.8.8", "8.8.4.4"]
  domain_name  = "iot.local"
}

# Create a VLAN-only network (no routing/DHCP)
resource "unifi_network" "vlan_only" {
  name    = "External VLAN"
  purpose = "vlan-only"
  vlan_id = 200
}
