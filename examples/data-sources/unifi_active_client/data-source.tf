# Look up an active client by MAC address
data "unifi_active_client" "by_mac" {
  mac = "aa:bb:cc:dd:ee:ff"
}

# Look up an active client by display name
data "unifi_active_client" "by_name" {
  display_name = "My Laptop"
}
