# Look up a static DNS record by key (hostname)
data "unifi_static_dns" "example" {
  key = "server.local"
}

# Or look up by ID
data "unifi_static_dns" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "dns_value" {
  value = data.unifi_static_dns.example.value
}
