resource "time_offset" "in_a_day" {
  offset_days = 1
}

resource "meltcloud_enrollment_image" "example" {
  name                = "my-image"
  expires_at          = time_offset.in_a_day.rfc3339
  install_disk_device = "/dev/vda"
  vlan                = 100
}