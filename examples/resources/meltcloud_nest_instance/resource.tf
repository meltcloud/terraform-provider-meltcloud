# add a second and a third nest instance, each in a zone of its own
resource "meltcloud_nest_instance" "nest_1" {
  name               = "nest-1"
  zone_id            = meltcloud_zone.zrh_dc2.id
  network_profile_id = meltcloud_network_profile.nest.id

  ip_addresses = [
    {
      ip_pool_id = meltcloud_ip_pool.hosts.id
      address    = "10.0.10.11"
    },
  ]

  boot_disk = {
    device        = "/dev/disk/by-path/pci-0000:00:17.0-ata-1"
    mirror        = true
    mirror_device = "/dev/disk/by-path/pci-0000:00:17.0-ata-2"
  }

  etcd_disk = {
    enabled = true
    device  = "/dev/disk/by-path/pci-0000:00:17.0-ata-3"
  }
}

# on a nest that uses BGP, each instance peers with its routers
resource "meltcloud_nest_instance" "nest_2" {
  name               = "nest-2"
  zone_id            = meltcloud_zone.zrh_dc3.id
  network_profile_id = meltcloud_network_profile.nest.id

  ip_addresses = [
    {
      ip_pool_id = meltcloud_ip_pool.hosts.id
      address    = "10.0.10.12"
    },
  ]

  boot_disk = {
    device = "/dev/disk/by-path/pci-0000:00:17.0-ata-1"
  }

  bgp = {
    local_asn = 65012
    peers = [
      {
        ip                  = "10.0.10.1"
        asn                 = 65000
        password_wo         = var.bgp_password
        password_wo_version = 1
        bfd = {
          enabled = true
        }
      },
    ]
  }
}
