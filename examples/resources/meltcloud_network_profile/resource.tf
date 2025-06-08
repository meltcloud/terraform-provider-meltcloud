resource "meltcloud_network_profile" "example" {
  name = "profile1"

  vlan {
    vlan      = 1
    dhcp      = false
    interface = "workload"
  }

  vlan {
    vlan      = 2
    dhcp      = false
    interface = "storage"
  }

  bridge {
    name      = "workload.1"
    interface = "br.workload"
    dhcp      = true
  }

  bridge {
    name      = "storage.2"
    interface = "br.storage"
    dhcp      = true
  }

  bond {
    name       = "workload"
    kind       = "default"
    dhcp       = true
    interfaces = "eth0,eth1"
  }

  bond {
    name       = "storage"
    kind       = "lacp"
    dhcp       = false
    interfaces = "eth2,eth3"
  }
}