# Look up a traffic rule by description
data "unifi_traffic_rule" "example" {
  description = "Block Social Media"
}

# Or look up by ID
data "unifi_traffic_rule" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "rule_action" {
  value = data.unifi_traffic_rule.example.action
}
