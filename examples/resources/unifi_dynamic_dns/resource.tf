# DuckDNS configuration
resource "unifi_dynamic_dns" "duckdns" {
  service  = "duckdns"
  hostname = "mynetwork.duckdns.org"
  password = "your-duckdns-token"
}

# Cloudflare configuration
resource "unifi_dynamic_dns" "cloudflare" {
  service  = "cloudflare"
  hostname = "home.example.com"
  login    = "email@example.com"
  password = "cloudflare-api-token"
}

# No-IP configuration
resource "unifi_dynamic_dns" "noip" {
  service  = "noip"
  hostname = "myhost.no-ip.org"
  login    = "username"
  password = "password"
}

# Custom DDNS provider
resource "unifi_dynamic_dns" "custom" {
  service  = "custom"
  hostname = "home.example.com"
  server   = "update.example.com"
  login    = "username"
  password = "password"
  options  = "myip=<ip>&hostname=<h>"
}

# Using WAN2 interface
resource "unifi_dynamic_dns" "wan2" {
  service   = "dyndns"
  hostname  = "backup.example.com"
  login     = "username"
  password  = "password"
  interface = "wan2"
}
