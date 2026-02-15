# Create a DHCP reservation for an ESXi host
resource "unifi_user" "esxi_host" {
  mac          = "aa:bb:cc:dd:ee:01"
  name         = "ESXi Host 1"
  use_fixed_ip = true
  fixed_ip     = "192.168.1.100"
  network_id   = unifi_network.servers.id
}

# Create a user with local DNS record
resource "unifi_user" "nas" {
  mac                      = "aa:bb:cc:dd:ee:02"
  name                     = "Synology NAS"
  use_fixed_ip             = true
  fixed_ip                 = "192.168.1.50"
  network_id               = unifi_network.servers.id
  local_dns_record         = "nas"
  local_dns_record_enabled = true
}

# Block a device from network access
resource "unifi_user" "blocked_device" {
  mac     = "aa:bb:cc:dd:ee:03"
  name    = "Untrusted Device"
  blocked = true
}

# Assign a device to a bandwidth-limited user group
resource "unifi_user" "iot_device" {
  mac          = "aa:bb:cc:dd:ee:04"
  name         = "Smart Thermostat"
  usergroup_id = unifi_user_group.iot_throttled.id
  note         = "Living room thermostat"
  noted        = true
}
