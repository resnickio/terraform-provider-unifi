# Look up a firewall zone by name
data "unifi_firewall_zone" "example" {
  name = "Internal"
}

# Or look up by ID
data "unifi_firewall_zone" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source in a firewall policy
resource "unifi_firewall_policy" "example" {
  name        = "Example Policy"
  source_zone = data.unifi_firewall_zone.example.id
  # ...
}
