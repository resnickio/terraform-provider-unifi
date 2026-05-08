# Block specific domains
resource "unifi_traffic_rule" "block_social" {
  name            = "Block-Social-Media"
  action          = "BLOCK"
  matching_target = "DOMAIN"

  domains = [
    {
      domain      = "*.facebook.com"
      description = "Facebook"
    },
    {
      domain      = "*.twitter.com"
      description = "Twitter"
    },
    {
      domain = "*.tiktok.com"
    },
  ]

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

# Block during work hours with schedule
resource "unifi_traffic_rule" "work_hours_block" {
  name             = "Block-Gaming-Work-Hours"
  action           = "BLOCK"
  matching_target  = "APP"
  app_category_ids = ["gaming"]

  schedule = {
    mode             = "EVERY_WEEK"
    time_range_start = "09:00"
    time_range_end   = "17:00"
    repeat_on_days   = ["mon", "tue", "wed", "thu", "fri"]
  }

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

# Bandwidth limit for streaming
resource "unifi_traffic_rule" "limit_streaming" {
  name             = "Limit-Streaming"
  action           = "ALLOW"
  matching_target  = "APP"
  app_category_ids = ["streaming_media"]

  bandwidth_limit = {
    download_limit_kbps = 25000
    upload_limit_kbps   = 5000
    enabled             = true
  }

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

# Block specific IP ranges
resource "unifi_traffic_rule" "block_ips" {
  name            = "Block-Suspicious-IPs"
  action          = "BLOCK"
  matching_target = "IP"
  ip_addresses    = ["192.168.100.0/24", "10.10.10.0/24"]

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

# Allow specific regions
resource "unifi_traffic_rule" "allow_regions" {
  name            = "Allow-US-Only"
  action          = "ALLOW"
  matching_target = "REGION"
  regions         = ["US", "CA"]

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

# Disabled rule (for maintenance)
resource "unifi_traffic_rule" "disabled" {
  name        = "Maintenance-Rule"
  action      = "BLOCK"
  enabled     = false
  description = "Disabled for maintenance"

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}

# Target specific devices
resource "unifi_traffic_rule" "device_specific" {
  name            = "Kids-Device-Block"
  action          = "BLOCK"
  matching_target = "INTERNET"

  target_devices = [
    {
      type       = "CLIENT"
      client_mac = "aa:bb:cc:dd:ee:ff"
    },
    {
      type       = "NETWORK"
      network_id = unifi_network.kids.id
    },
  ]
}

# Apply to multiple networks
resource "unifi_traffic_rule" "multi_network" {
  name            = "VLAN-Block"
  action          = "BLOCK"
  matching_target = "INTERNET"
  network_ids     = [unifi_network.guest.id, unifi_network.iot.id]

  target_devices = [{
    type = "ALL_CLIENTS"
  }]
}
