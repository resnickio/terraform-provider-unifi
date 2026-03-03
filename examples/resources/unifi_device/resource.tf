# Manage an adopted device's settings
resource "unifi_device" "switch" {
  mac  = "aa:bb:cc:dd:ee:01"
  name = "Office Switch"
}

# Manage device with LED override
resource "unifi_device" "ap" {
  mac          = "aa:bb:cc:dd:ee:02"
  name         = "Lobby AP"
  led_override = "off"
}
