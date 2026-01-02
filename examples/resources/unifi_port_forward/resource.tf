# Forward HTTPS traffic to a web server
resource "unifi_port_forward" "web_server" {
  name           = "Web Server HTTPS"
  enabled        = true
  protocol       = "tcp"
  dst_port       = "443"
  fwd_port       = "443"
  fwd_ip         = "10.0.100.50"
  pfwd_interface = "wan"
  log            = false
}

# Forward a range of game server ports
resource "unifi_port_forward" "game_server" {
  name           = "Game Server"
  enabled        = true
  protocol       = "udp"
  dst_port       = "27015-27020"
  fwd_port       = "27015-27020"
  fwd_ip         = "10.0.100.60"
  pfwd_interface = "wan"
}

# Port forward with source restriction
resource "unifi_port_forward" "ssh_restricted" {
  name           = "SSH (Restricted)"
  enabled        = true
  protocol       = "tcp"
  dst_port       = "22"
  fwd_port       = "22"
  fwd_ip         = "10.0.100.10"
  src            = "203.0.113.0/24"
  pfwd_interface = "wan"
  log            = true
}
