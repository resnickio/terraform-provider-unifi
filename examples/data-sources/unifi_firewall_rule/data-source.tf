# Look up a legacy firewall rule by name
data "unifi_firewall_rule" "example" {
  name = "Block Telnet"
}

# Or look up by ID
data "unifi_firewall_rule" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "rule_action" {
  value = data.unifi_firewall_rule.example.action
}
