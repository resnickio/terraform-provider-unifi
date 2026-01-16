# Look up a port profile by name
data "unifi_port_profile" "example" {
  name = "All"
}

# Or look up by ID
data "unifi_port_profile" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "profile_native_network" {
  value = data.unifi_port_profile.example.native_network_id
}
