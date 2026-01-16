# Look up a network by name
data "unifi_network" "example" {
  name = "Default"
}

# Or look up by ID
data "unifi_network" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "network_vlan" {
  value = data.unifi_network.example.vlan_id
}
