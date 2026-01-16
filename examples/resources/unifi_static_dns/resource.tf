# Basic A record
resource "unifi_static_dns" "server" {
  key         = "server.internal.example.com"
  value       = "192.168.1.100"
  record_type = "A"
}

# A record with TTL
resource "unifi_static_dns" "db" {
  key         = "db.internal.example.com"
  value       = "192.168.1.101"
  record_type = "A"
  ttl         = 300
}

# CNAME record
resource "unifi_static_dns" "alias" {
  key         = "api.internal.example.com"
  value       = "server.internal.example.com"
  record_type = "CNAME"
}

# MX record with priority
resource "unifi_static_dns" "mail" {
  key         = "example.com"
  value       = "mail.example.com"
  record_type = "MX"
  priority    = 10
}

# SRV record
resource "unifi_static_dns" "sip" {
  key         = "_sip._tcp.example.com"
  value       = "sipserver.example.com"
  record_type = "SRV"
  port        = 5060
  priority    = 10
  weight      = 5
}

# Disabled record
resource "unifi_static_dns" "legacy" {
  key         = "oldserver.internal.example.com"
  value       = "192.168.1.50"
  record_type = "A"
  enabled     = false
}
