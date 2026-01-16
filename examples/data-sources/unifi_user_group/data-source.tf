# Look up a user group by name
data "unifi_user_group" "example" {
  name = "Default"
}

# Or look up by ID
data "unifi_user_group" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "group_qos_rate" {
  value = data.unifi_user_group.example.qos_rate_max_down
}
