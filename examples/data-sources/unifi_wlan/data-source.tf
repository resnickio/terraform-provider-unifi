# Look up a WLAN by name
data "unifi_wlan" "example" {
  name = "Corporate WiFi"
}

# Or look up by ID
data "unifi_wlan" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "wlan_security" {
  value = data.unifi_wlan.example.security
}
