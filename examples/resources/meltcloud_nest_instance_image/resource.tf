resource "meltcloud_nest_instance_image" "nest_1" {
  nest_instance_id = meltcloud_nest_instance.nest_1.id
  name             = "nest-1"
  expires_at       = "2026-12-31T23:59:59Z"
}

# boot the machine from the ISO, for example through its BMC
output "nest_1_iso" {
  value     = meltcloud_nest_instance_image.nest_1.https_url_iso_amd64
  sensitive = true
}
