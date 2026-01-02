# Create a secure WPA2 wireless network
resource "unifi_wlan" "main" {
  name         = "MyNetwork"
  security     = "wpapsk"
  wpa_mode     = "wpa2"
  passphrase   = var.wifi_password
  ap_group_ids = [data.unifi_ap_group.default.id]

  # Optional: Associate with a specific network/VLAN
  network_id = unifi_network.main.id
}

# Create a guest network with isolation
resource "unifi_wlan" "guest" {
  name         = "Guest Network"
  security     = "wpapsk"
  wpa_mode     = "wpa2"
  passphrase   = var.guest_wifi_password
  ap_group_ids = [data.unifi_ap_group.default.id]

  is_guest     = true
  l2_isolation = true
}

# Create a WPA3 enabled network
resource "unifi_wlan" "secure" {
  name          = "Secure Network"
  security      = "wpapsk"
  wpa_mode      = "wpa2"
  wpa3_support  = true
  wpa3_transition = true
  passphrase    = var.secure_wifi_password
  ap_group_ids  = [data.unifi_ap_group.default.id]

  pmf_mode = "required"
}

variable "wifi_password" {
  type      = string
  sensitive = true
}

variable "guest_wifi_password" {
  type      = string
  sensitive = true
}

variable "secure_wifi_password" {
  type      = string
  sensitive = true
}

resource "unifi_network" "main" {
  name    = "Main Network"
  purpose = "corporate"
  vlan_id = 1
}
