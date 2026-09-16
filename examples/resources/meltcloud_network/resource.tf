# stretched L2: one subnet, used by machines on both sites
resource "meltcloud_network" "stretched" {
  name = "wl-prd"
}

resource "meltcloud_subnet" "stretched" {
  network_id = meltcloud_network.stretched.id
  name       = "wl-prd"
  addressing = "dhcp"
  vlan       = 10
}

# subnet per rack or zone: one subnet per segment, in the same network
resource "meltcloud_network" "per_zone" {
  name = "wl-prd-zoned"
}

resource "meltcloud_subnet" "az1" {
  network_id = meltcloud_network.per_zone.id
  name       = "wl-prd-az1"
  addressing = "dhcp"
  vlan       = 10
}

resource "meltcloud_subnet" "az2" {
  network_id = meltcloud_network.per_zone.id
  name       = "wl-prd-az2"
  addressing = "dhcp"
  vlan       = 20
}
