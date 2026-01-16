# Look up a NAT rule by description
data "unifi_nat_rule" "example" {
  description = "Web Server DNAT"
}

# Or look up by ID
data "unifi_nat_rule" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "nat_type" {
  value = data.unifi_nat_rule.example.type
}
