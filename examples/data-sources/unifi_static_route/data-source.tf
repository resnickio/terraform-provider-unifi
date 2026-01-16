# Look up a static route by name
data "unifi_static_route" "example" {
  name = "VPN Route"
}

# Or look up by ID
data "unifi_static_route" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "route_network" {
  value = data.unifi_static_route.example.network
}
