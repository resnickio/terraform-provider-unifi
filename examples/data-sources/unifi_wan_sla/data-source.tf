# Look up a WAN SLA monitor by name
data "unifi_wan_sla" "primary" {
  name = "Primary WAN"
}
