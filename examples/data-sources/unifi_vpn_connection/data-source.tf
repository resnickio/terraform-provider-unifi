# Look up a VPN connection by name
data "unifi_vpn_connection" "site_to_site" {
  name = "Office VPN"
}
