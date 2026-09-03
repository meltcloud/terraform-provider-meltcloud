resource "time_offset" "in_a_day" {
  offset_days = 1
}

resource "meltcloud_enrollment_image" "example" {
  name                = "my-image"
  expires_at          = time_offset.in_a_day.rfc3339
  install_disk_device = "/dev/disk/by-path/pci-0000:00:17.0-ata-1"
  network_profile_id  = meltcloud_network_profile.example.id
}

# a mirrored install requires both disks to be named explicitly;
# by-path devices describe the slot, so one image works for a whole fleet of identical servers
resource "meltcloud_enrollment_image" "mirrored" {
  name                       = "my-mirrored-image"
  expires_at                 = time_offset.in_a_day.rfc3339
  install_disk_device        = "/dev/disk/by-path/pci-0000:00:17.0-ata-1"
  install_disk_mirror        = true
  install_disk_mirror_device = "/dev/disk/by-path/pci-0000:00:17.0-ata-2"
  network_profile_id         = meltcloud_network_profile.example.id
}
