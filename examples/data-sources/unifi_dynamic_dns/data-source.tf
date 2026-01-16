# Look up a dynamic DNS configuration by hostname
data "unifi_dynamic_dns" "example" {
  hostname = "home.example.com"
}

# Or look up by ID
data "unifi_dynamic_dns" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "ddns_service" {
  value = data.unifi_dynamic_dns.example.service
}
