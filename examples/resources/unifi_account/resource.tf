# Create a RADIUS account for 802.1X authentication
resource "unifi_account" "wireless_user" {
  name       = "jsmith"
  x_password = "secure-password-123"
}

# Create a RADIUS account with VLAN assignment
resource "unifi_account" "vlan_user" {
  name       = "guest-user"
  x_password = "guest-pass-456"
  vlan       = 100
}
