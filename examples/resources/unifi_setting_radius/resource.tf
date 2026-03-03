# Configure site RADIUS server settings
resource "unifi_setting_radius" "example" {
  enabled            = true
  accounting_enabled = true
  auth_port          = 1812
  acct_port          = 1813
  x_secret           = "radius-shared-secret"
  tunneled_reply     = true
}
