# Block IoT devices from accessing the LAN
resource "unifi_firewall_rule" "block_iot_to_lan" {
  name       = "Block IoT to LAN"
  ruleset    = "LAN_IN"
  action     = "drop"
  rule_index = 2000
  enabled    = true

  src_firewall_group_ids = [unifi_firewall_group.iot_devices.id]

  logging = true
}

# Allow specific ports from WAN
resource "unifi_firewall_rule" "allow_web" {
  name       = "Allow Web Traffic"
  ruleset    = "WAN_IN"
  action     = "accept"
  rule_index = 2001
  protocol   = "tcp"

  dst_port = "443"
}

resource "unifi_firewall_group" "iot_devices" {
  name       = "IoT Devices"
  group_type = "address-group"
  members    = ["10.0.100.0/24"]
}
