# Block specific domains
resource "unifi_traffic_rule" "block_social" {
  name            = "Block-Social-Media"
  action          = "BLOCK"
  matching_target = "DOMAIN"

  domains {
    domain      = "*.facebook.com"
    description = "Facebook"
  }

  domains {
    domain      = "*.twitter.com"
    description = "Twitter"
  }

  domains {
    domain = "*.tiktok.com"
  }
}

# Block during work hours with schedule
resource "unifi_traffic_rule" "work_hours_block" {
  name            = "Block-Gaming-Work-Hours"
  action          = "BLOCK"
  matching_target = "APP"
  app_category_ids = ["gaming"]

  schedule {
    mode             = "CUSTOM"
    time_range_start = "09:00"
    time_range_end   = "17:00"
    days_of_week     = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"]
  }
}

# Bandwidth limit for streaming
resource "unifi_traffic_rule" "limit_streaming" {
  name            = "Limit-Streaming"
  action          = "ALLOW"
  matching_target = "APP"
  app_category_ids = ["streaming_media"]

  bandwidth_limit {
    download_limit_kbps = 25000
    upload_limit_kbps   = 5000
    enabled             = true
  }
}

# Block specific IP ranges
resource "unifi_traffic_rule" "block_ips" {
  name            = "Block-Suspicious-IPs"
  action          = "BLOCK"
  matching_target = "IP"
  ip_addresses    = ["192.168.100.0/24", "10.10.10.0/24"]
}

# Allow specific regions
resource "unifi_traffic_rule" "allow_regions" {
  name            = "Allow-US-Only"
  action          = "ALLOW"
  matching_target = "REGION"
  regions         = ["US", "CA"]
}

# Disabled rule (for maintenance)
resource "unifi_traffic_rule" "disabled" {
  name        = "Maintenance-Rule"
  action      = "BLOCK"
  enabled     = false
  description = "Disabled for maintenance"
}

# Target specific devices
resource "unifi_traffic_rule" "device_specific" {
  name   = "Kids-Device-Block"
  action = "BLOCK"

  target_devices {
    type       = "CLIENT"
    client_mac = "aa:bb:cc:dd:ee:ff"
  }

  target_devices {
    type       = "NETWORK"
    network_id = unifi_network.kids.id
  }

  matching_target = "INTERNET"
}
