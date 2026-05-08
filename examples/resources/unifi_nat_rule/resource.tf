# NOTE: The UniFi v2 NAT API on current controllers does not accept the
# carrier fields (source_address, source_port, dest_address, dest_port,
# translated_ip, translated_port). The provider currently manages only the
# rule shell. Manage NAT translation rules via the UniFi UI until the
# upstream API stabilizes.

# Masquerade NAT shell
resource "unifi_nat_rule" "masquerade" {
  type        = "MASQUERADE"
  description = "Masquerade rule"
}

# Logged rule shell
resource "unifi_nat_rule" "logged" {
  type        = "MASQUERADE"
  description = "NAT with logging enabled"
  logging     = true
}

# Disabled rule shell
resource "unifi_nat_rule" "disabled" {
  type        = "MASQUERADE"
  description = "Disabled NAT rule"
  enabled     = false
}
