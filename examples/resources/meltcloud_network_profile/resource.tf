# one interface, found by the machine itself, addressed on the untagged segment
resource "meltcloud_network_profile" "example" {
  name = "profile1"

  uplink {
    name = "up0"
    mode = "auto"

    host_network {
      subnet_id   = meltcloud_subnet.mgmt.id
      vlan_tagged = false
      primary     = true
    }
  }
}

# two interfaces bonded, carrying the management segment untagged and storage on a tagged VLAN
resource "meltcloud_network_profile" "bonded" {
  name = "profile2"

  uplink {
    name       = "up0"
    mode       = "bond"
    identifier = "kernel_name"
    interfaces = ["eth0", "eth1"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.mgmt.id
      vlan_tagged = false
      primary     = true
    }

    host_network {
      subnet_id   = meltcloud_subnet.storage.id
      vlan_tagged = true
      primary     = false
    }
  }
}
