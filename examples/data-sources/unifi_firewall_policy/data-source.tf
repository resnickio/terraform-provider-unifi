# Look up a firewall policy by name
data "unifi_firewall_policy" "example" {
  name = "LAN to WAN"
}

# Or look up by ID
data "unifi_firewall_policy" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "policy_action" {
  value = data.unifi_firewall_policy.example.action
}
