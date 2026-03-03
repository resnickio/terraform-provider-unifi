# Configure site management settings
resource "unifi_setting_mgmt" "example" {
  auto_upgrade = true
  led_enabled  = true
  alert_enabled = true
  x_ssh_enabled = true
  x_ssh_auth_password_enabled = true
  x_ssh_username = "admin"
  x_ssh_password = "secure-ssh-pass"
}
