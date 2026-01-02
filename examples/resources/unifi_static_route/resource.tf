# Create a static route to a remote network via gateway
resource "unifi_static_route" "remote_office" {
  name                  = "Remote Office"
  enabled               = true
  type                  = "static-route"
  static_route_network  = "192.168.10.0/24"
  static_route_nexthop  = "10.0.1.254"
  static_route_distance = 1
  static_route_type     = "nexthop-route"
}

# Create a blackhole route to drop traffic
resource "unifi_static_route" "blackhole" {
  name                 = "Block Bad Network"
  enabled              = true
  type                 = "static-route"
  static_route_network = "10.99.0.0/16"
  static_route_type    = "blackhole"
}
