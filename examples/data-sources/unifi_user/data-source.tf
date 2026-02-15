# Look up a user by MAC address
data "unifi_user" "example" {
  mac = "aa:bb:cc:dd:ee:01"
}

# Or look up by ID
data "unifi_user" "by_id" {
  id = "60a1b2c3d4e5f67890123456"
}

# Use the data source
output "user_fixed_ip" {
  value = data.unifi_user.example.fixed_ip
}
