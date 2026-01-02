# Create firewall zones
resource "unifi_firewall_zone" "iot" {
  name = "IoT Zone"
}

resource "unifi_firewall_zone" "servers" {
  name = "Server Zone"
}

# Block IoT zone from accessing Server zone
resource "unifi_firewall_policy" "block_iot_to_servers" {
  name    = "Block IoT to Servers"
  enabled = true
  action  = "BLOCK"
  index   = 1000

  source {
    zone_id = unifi_firewall_zone.iot.id
  }

  destination {
    zone_id = unifi_firewall_zone.servers.id
  }

  logging = true
}

# Allow HTTPS traffic from any zone
resource "unifi_firewall_policy" "allow_https" {
  name      = "Allow HTTPS"
  enabled   = true
  action    = "ALLOW"
  protocol  = "tcp"
  index     = 1001

  destination {
    port = "443"
  }
}
