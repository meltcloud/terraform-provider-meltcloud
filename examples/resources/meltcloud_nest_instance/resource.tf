# add a second and a third nest instance, each in a zone of its own
resource "meltcloud_nest_instance" "nest_1" {
  name               = "nest-1"
  zone_id            = meltcloud_zone.zrh_dc2.id
  network_profile_id = meltcloud_network_profile.nest.id

  ip_addresses = [
    { ip_pool_id = meltcloud_ip_pool.hosts.id, address = "10.0.10.11" },
  ]

  install_disk_device = "/dev/disk/by-id/nvme-SAMSUNG_MZQL2960HCJR-00A07_S64FNE0R123456"
}

# on a nest that uses BGP, each instance peers with its routers
resource "meltcloud_nest_instance" "nest_2" {
  name               = "nest-2"
  zone_id            = meltcloud_zone.zrh_dc3.id
  network_profile_id = meltcloud_network_profile.nest.id

  ip_addresses = [
    { ip_pool_id = meltcloud_ip_pool.hosts.id, address = "10.0.10.12" },
  ]

  bgp = {
    local_asn = 65012
    peers = [
      { ip = "10.0.10.1", asn = 65000, password = var.bgp_password, bfd = {} },
    ]
  }
}
