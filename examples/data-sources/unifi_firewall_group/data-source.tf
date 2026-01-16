# Look up a firewall group by name
data "unifi_firewall_group" "example" {
  name = "Blocked IPs"
}

# Or look up by ID
data "unifi_firewall_group" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "group_members" {
  value = data.unifi_firewall_group.example.members
}
