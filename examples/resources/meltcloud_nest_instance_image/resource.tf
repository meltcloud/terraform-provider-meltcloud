resource "time_offset" "in_a_day" {
  offset_days = 1
}

resource "meltcloud_nest_instance_image" "nest_1" {
  nest_instance_id = meltcloud_nest_instance.nest_1.id
  name             = "nest-1"
  expires_at       = time_offset.in_a_day.rfc3339
}

# boot the host from the ISO, for example through its BMC
output "nest_1_iso" {
  value     = meltcloud_nest_instance_image.nest_1.https_url_iso_amd64
  sensitive = true
}
