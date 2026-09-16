# a DHCP server on the segment hands out the addresses
resource "meltcloud_subnet" "workload_dhcp" {
  network_id = meltcloud_network.example.id
  name       = "workload"
  addressing = "dhcp"

  # settings to override, or to add where the DHCP server delivers none
  # example: MTU not set by DHCP, let's set it for jumbo frames:
  mtu = 9000
}

# DHCP with VLAN specified
resource "meltcloud_subnet" "storage_dhcp" {
  network_id = meltcloud_network.example.id
  name       = "storage"
  addressing = "dhcp"

  vlan = 300

  # settings to override, or to add where the DHCP server delivers none
  # example: additional routes, not provided by dhcp
  route {
    destination = "10.30.0.0/16"
    via         = "10.20.0.254"
    metric      = 200
  }
}

# an IPAM subnet
resource "meltcloud_subnet" "workload_ipam" {
  network_id = meltcloud_network.example.id
  name       = "workload"
  addressing = "ipam"

  ip_pool_id = meltcloud_ip_pool.example.id
  gateway    = "10.20.0.1"
  dns        = ["10.20.0.53", "10.20.0.54"]
  ntp        = ["10.20.0.60"]
  domains    = ["lab.example.com"]
}

# a tagged segment addressed by meltcloud, with its own pool
resource "meltcloud_subnet" "storage_ipam" {
  network_id = meltcloud_network.example.id
  name       = "storage"
  addressing = "ipam"

  ip_pool_id = meltcloud_ip_pool.storage.id
  vlan       = 300
  gateway    = "10.30.0.1"
  dns        = ["10.20.0.53", "10.20.0.54"]
  mtu        = 9000 # jumbo frames
}
