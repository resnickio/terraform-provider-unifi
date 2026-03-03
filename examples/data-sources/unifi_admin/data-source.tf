data "unifi_admin" "example" {
  name = "admin"
}

output "admin_email" {
  value = data.unifi_admin.example.email
}
