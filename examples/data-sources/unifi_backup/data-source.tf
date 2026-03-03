data "unifi_backup" "all" {}

output "backup_count" {
  value = length(data.unifi_backup.all.backups)
}
