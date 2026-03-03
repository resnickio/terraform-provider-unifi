resource "unifi_content_filtering" "example" {
  enabled         = true
  blocked_domains = ["malware.example.com"]
  allowed_domains = ["safe.example.com"]
}
