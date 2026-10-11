# add a second and a third nest instance, each in a zone of its own
resource "meltcloud_nest_instance" "nest_1" {
  name    = "nest-1"
  ip      = "10.0.10.11"
  zone_id = meltcloud_zone.dc2.id

  install_disk_device = "/dev/disk/by-id/nvme-SAMSUNG_MZQL2960HCJR-00A07_S64FNE0R123456"
}

resource "meltcloud_nest_instance" "nest_2" {
  name    = "nest-2"
  ip      = "10.0.10.12"
  zone_id = meltcloud_zone.dc3.id

  # a nest adds one instance at a time
  depends_on = [meltcloud_nest_instance.nest_1]
}

# boot each machine from its ISO, for example through its BMC
output "nest_1_iso" {
  value     = meltcloud_nest_instance.nest_1.url_iso_amd64
  sensitive = true
}
