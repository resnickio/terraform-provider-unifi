# Basic RADIUS profile
resource "unifi_radius_profile" "basic" {
  name = "Corporate RADIUS"
}

# Full RADIUS profile with auth and accounting servers
resource "unifi_radius_profile" "full" {
  name = "Enterprise RADIUS"

  use_usg_auth_server = false
  use_usg_acct_server = false

  vlan_enabled   = true
  vlan_wlan_mode = "optional"

  interim_update_enabled  = true
  interim_update_interval = 600

  auth_server {
    ip     = "10.0.0.100"
    port   = 1812
    secret = "auth-secret-here"
  }

  acct_server {
    ip     = "10.0.0.100"
    port   = 1813
    secret = "acct-secret-here"
  }
}

# RADIUS profile with multiple authentication servers for redundancy
resource "unifi_radius_profile" "redundant" {
  name = "Redundant RADIUS"

  auth_server {
    ip     = "10.0.0.100"
    port   = 1812
    secret = "primary-secret"
  }

  auth_server {
    ip     = "10.0.0.101"
    port   = 1812
    secret = "secondary-secret"
  }
}
