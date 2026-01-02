# Create a custom firewall zone for IoT devices
resource "unifi_firewall_zone" "iot" {
  name        = "IoT Zone"
  network_ids = [unifi_network.iot.id]
}

# Create a zone for servers
resource "unifi_firewall_zone" "servers" {
  name        = "Server Zone"
  network_ids = [unifi_network.servers.id]
}

# Supporting network resources
resource "unifi_network" "iot" {
  name    = "IoT Network"
  purpose = "corporate"
  vlan_id = 100
  subnet  = "10.0.100.1/24"
}

resource "unifi_network" "servers" {
  name    = "Server Network"
  purpose = "corporate"
  vlan_id = 200
  subnet  = "10.0.200.1/24"
}
