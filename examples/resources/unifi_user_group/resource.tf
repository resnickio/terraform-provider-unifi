# Create a bandwidth-limited user group for guests
resource "unifi_user_group" "guest_limited" {
  name             = "Guest Limited"
  qos_rate_max_down = 10000  # 10 Mbps download
  qos_rate_max_up   = 5000   # 5 Mbps upload
}

# Create an unlimited user group
resource "unifi_user_group" "unlimited" {
  name             = "Unlimited"
  qos_rate_max_down = -1  # Unlimited
  qos_rate_max_up   = -1  # Unlimited
}

# Create a heavily throttled IoT group
resource "unifi_user_group" "iot_throttled" {
  name             = "IoT Throttled"
  qos_rate_max_down = 1000  # 1 Mbps download
  qos_rate_max_up   = 500   # 500 Kbps upload
}
