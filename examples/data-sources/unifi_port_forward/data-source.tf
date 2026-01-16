# Look up a port forward by name
data "unifi_port_forward" "example" {
  name = "SSH Server"
}

# Or look up by ID
data "unifi_port_forward" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "forward_destination" {
  value = data.unifi_port_forward.example.destination_ip
}
