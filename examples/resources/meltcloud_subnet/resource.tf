# a DHCP server on the segment hands out the addresses
resource "meltcloud_subnet" "mgmt" {
  network_id = meltcloud_network.example.id
  name       = "mgmt"
  addressing = "dhcp"

  # settings to override, or to add where the DHCP server delivers none
  mtu = 9000
}

# a tagged segment, still addressed by DHCP
resource "meltcloud_subnet" "storage" {
  network_id = meltcloud_network.example.id
  name       = "storage"
  addressing = "dhcp"
  vlan       = 300

  # settings to override, or to add where the DHCP server delivers none
  route {
    destination = "10.30.0.0/16"
    via         = "10.20.0.254"
    metric      = 200
  }
}

# meltcloud hands out the addresses & network settings that a DHCP server would
resource "meltcloud_subnet" "wl" {
  network_id = meltcloud_network.example.id
  name       = "wl"
  addressing = "ipam"
  ip_pool_id = meltcloud_ip_pool.example.id
  gateway    = "10.20.0.1"
  dns        = ["10.20.0.53"]
}
