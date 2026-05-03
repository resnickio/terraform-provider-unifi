# Look up an AP group by name
data "unifi_ap_group" "example" {
  name = "Default"
}

# Or look up by ID
data "unifi_ap_group" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "ap_group_id" {
  value = data.unifi_ap_group.example.id
}
