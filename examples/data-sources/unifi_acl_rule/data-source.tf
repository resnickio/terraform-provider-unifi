# Look up an ACL rule by name
data "unifi_acl_rule" "example" {
  name = "Block LAN to WLAN Multicast and Broadcast"
}
