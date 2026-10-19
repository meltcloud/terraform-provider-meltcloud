data "meltcloud_foundry" "this" {}

output "foundry_version" {
  value = data.meltcloud_foundry.this.version
}
