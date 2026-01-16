# Basic port profile with just a name
resource "unifi_port_profile" "basic" {
  name = "Basic"
}

# Port profile with native VLAN
resource "unifi_port_profile" "workstation" {
  name              = "Workstation"
  native_network_id = unifi_network.corporate.id
}

# VLAN trunk allowing all VLANs
resource "unifi_port_profile" "trunk_all" {
  name              = "Trunk-All"
  native_network_id = unifi_network.management.id
  tagged_vlan_mgmt  = "all"
}

# Custom trunk - allows all VLANs EXCEPT specified ones
# Use excluded_network_ids to specify which networks to block
resource "unifi_port_profile" "hypervisor" {
  name              = "Hypervisor"
  native_network_id = unifi_network.management.id
  tagged_vlan_mgmt  = "custom"
  excluded_network_ids = [
    unifi_network.guest.id,
    unifi_network.iot.id,
  ]
}

# Access port - blocks all tagged VLANs
resource "unifi_port_profile" "access_only" {
  name              = "Access-Only"
  native_network_id = unifi_network.corporate.id
  tagged_vlan_mgmt  = "block"
}

# PoE camera with port isolation
resource "unifi_port_profile" "security_camera" {
  name              = "Security-Camera"
  native_network_id = unifi_network.security.id
  poe_mode          = "auto"
  isolation         = true
}

# Profile with storm control
resource "unifi_port_profile" "storm_protected" {
  name                    = "Storm-Protected"
  native_network_id       = unifi_network.servers.id
  stormctrl_bcast_enabled = true
  stormctrl_bcast_rate    = 1000
  stormctrl_mcast_enabled = true
  stormctrl_mcast_rate    = 1000
}
