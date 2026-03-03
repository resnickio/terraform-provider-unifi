resource "unifi_setting_guest_access" "example" {
  portal_enabled          = true
  portal_customized       = true
  auth                    = "none"
  portal_customized_title = "Welcome to Our Network"
}
