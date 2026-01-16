# Masquerade NAT for internal network
resource "unifi_nat_rule" "masquerade" {
  type           = "MASQUERADE"
  description    = "NAT for internal network"
  source_address = "192.168.1.0/24"
}

# DNAT for web server
resource "unifi_nat_rule" "web_server" {
  type            = "DNAT"
  description     = "Forward HTTP to web server"
  protocol        = "tcp"
  dest_port       = "80"
  translated_ip   = "192.168.1.100"
  translated_port = "8080"
}

# DNAT for HTTPS
resource "unifi_nat_rule" "https" {
  type            = "DNAT"
  description     = "Forward HTTPS to web server"
  protocol        = "tcp"
  dest_port       = "443"
  translated_ip   = "192.168.1.100"
  translated_port = "443"
}

# SNAT for specific source
resource "unifi_nat_rule" "snat_vpn" {
  type           = "SNAT"
  description    = "SNAT for VPN traffic"
  source_address = "10.10.0.0/16"
  translated_ip  = "192.168.1.1"
}

# NAT rule with logging
resource "unifi_nat_rule" "logged" {
  type        = "MASQUERADE"
  description = "NAT with logging enabled"
  logging     = true
}

# Disabled NAT rule
resource "unifi_nat_rule" "disabled" {
  type        = "DNAT"
  description = "Disabled NAT rule"
  enabled     = false
  dest_port   = "22"
}
