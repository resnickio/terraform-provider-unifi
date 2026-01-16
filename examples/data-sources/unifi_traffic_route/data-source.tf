# Look up a traffic route by description
data "unifi_traffic_route" "example" {
  description = "Work VPN Route"
}

# Or look up by ID
data "unifi_traffic_route" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "route_interface" {
  value = data.unifi_traffic_route.example.network_id
}
