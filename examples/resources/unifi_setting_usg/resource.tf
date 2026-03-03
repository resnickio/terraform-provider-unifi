# Configure site USG/gateway settings
resource "unifi_setting_usg" "example" {
  broadcast_ping = false
  mdns_enabled   = true
  upnp_enabled   = true
  upnp_secure_mode = true
  lldp_enable_all = true
}
