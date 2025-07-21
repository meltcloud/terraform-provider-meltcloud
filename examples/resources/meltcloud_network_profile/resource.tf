resource "meltcloud_network_profile" "example" {
  name = "profile1"

  link {
    name            = "link0"
    interfaces      = ["eth0", "eth1"]
    vlans           = []
    host_networking = false
    lacp            = true
    native_vlan     = false
  }

  link {
    name            = "link1"
    interfaces      = ["eth2"]
    vlans           = [300, 301]
    host_networking = true
    lacp            = false
    native_vlan     = true
  }
}