data "meltcloud_nest" "this" {}

output "nest_version" {
  value = data.meltcloud_nest.this.version
}
