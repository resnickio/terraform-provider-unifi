# Create an address group for IoT devices
resource "unifi_firewall_group" "iot_devices" {
  name       = "IoT Devices"
  group_type = "address-group"
  members = [
    "10.0.100.10",
    "10.0.100.11",
    "10.0.100.0/24",
  ]
}

# Create a port group for common web ports
resource "unifi_firewall_group" "web_ports" {
  name       = "Web Ports"
  group_type = "port-group"
  members    = ["80", "443", "8080-8090"]
}

# Create an IPv6 address group
resource "unifi_firewall_group" "ipv6_servers" {
  name       = "IPv6 Servers"
  group_type = "ipv6-address-group"
  members = [
    "2001:db8::1",
    "2001:db8::/32",
  ]
}
