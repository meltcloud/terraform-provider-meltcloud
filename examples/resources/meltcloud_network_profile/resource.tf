# single interface, auto-detected
resource "meltcloud_network_profile" "auto" {
  name = "workload"

  uplink {
    name = "workload"
    mode = "auto"

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }
  }
}

# single interface named, untagged vlan
resource "meltcloud_network_profile" "single_untagged" {
  name = "workload-single"

  uplink {
    name       = "workload"
    mode       = "single"
    identifier = "kernel_name"
    interfaces = ["eth0"]

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }
  }
}

# single interface named, tagged vlan
resource "meltcloud_network_profile" "single_tagged" {
  name = "workload-tagged"

  uplink {
    name       = "workload"
    mode       = "single"
    identifier = "kernel_name"
    interfaces = ["eth0"]

    host_network {
      subnet_id   = meltcloud_subnet.storage_dhcp.id
      vlan_tagged = true
      primary     = true
    }
  }
}

# two interfaces bonded, untagged vlan
resource "meltcloud_network_profile" "bonded_untagged" {
  name = "workload-bonded"

  uplink {
    name       = "workload"
    mode       = "bond"
    identifier = "kernel_name"
    interfaces = ["eth0", "eth1"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }
  }
}

# two interfaces bonded, tagged vlan
resource "meltcloud_network_profile" "bonded_tagged" {
  name = "storage-bonded"

  uplink {
    name       = "storage"
    mode       = "bond"
    identifier = "kernel_name"
    interfaces = ["eth0", "eth1"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.storage_dhcp.id
      vlan_tagged = true
      primary     = true
    }
  }
}

# single interface named, multiple subnets, all tagged
resource "meltcloud_network_profile" "single_multiple_tagged" {
  name = "main-single"

  uplink {
    name       = "main"
    mode       = "single"
    identifier = "kernel_name"
    interfaces = ["eth0"]

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = true
      primary     = true
    }

    host_network {
      subnet_id   = meltcloud_subnet.storage_dhcp.id
      vlan_tagged = true
      primary     = false
    }
  }
}

# single interface named, multiple subnets, one native and tagged subnets
resource "meltcloud_network_profile" "single_native_and_tagged" {
  name = "main-single-native"

  uplink {
    name       = "main"
    mode       = "single"
    identifier = "kernel_name"
    interfaces = ["eth0"]

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }

    host_network {
      subnet_id   = meltcloud_subnet.storage_dhcp.id
      vlan_tagged = true
      primary     = false
    }
  }
}

# one bond, multiple subnets, all tagged
resource "meltcloud_network_profile" "bonded_multiple_tagged" {
  name = "main-bonded-tagged"

  uplink {
    name       = "main"
    mode       = "bond"
    identifier = "kernel_name"
    interfaces = ["eth0", "eth1"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = true
      primary     = true
    }

    host_network {
      subnet_id   = meltcloud_subnet.storage_dhcp.id
      vlan_tagged = true
      primary     = false
    }
  }
}

# one bond, multiple subnets, one native and tagged subnets
resource "meltcloud_network_profile" "bonded_native_and_tagged" {
  name = "main"

  uplink {
    name       = "main"
    mode       = "bond"
    identifier = "kernel_name"
    interfaces = ["eth0", "eth1"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }

    host_network {
      subnet_id   = meltcloud_subnet.storage_dhcp.id
      vlan_tagged = true
      primary     = false
    }
  }
}

# multiple uplinks with a single interface each
resource "meltcloud_network_profile" "two_uplinks" {
  name = "workload-storage"

  uplink {
    name       = "workload"
    mode       = "single"
    identifier = "kernel_name"
    interfaces = ["eth0"]

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }
  }

  uplink {
    name       = "storage"
    mode       = "single"
    identifier = "kernel_name"
    interfaces = ["eth1"]

    host_network {
      subnet_id   = meltcloud_subnet.storage_ipam.id
      vlan_tagged = false
      primary     = false
    }
  }
}

# multiple uplinks with bonds each
resource "meltcloud_network_profile" "two_uplinks_bonded" {
  name = "workload-storage-bonded"

  uplink {
    name       = "workload"
    mode       = "bond"
    identifier = "mac_address"
    interfaces = ["00:1b:21:3c:4d:5e", "00:1b:21:3c:4d:5f"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.workload_dhcp.id
      vlan_tagged = false
      primary     = true
    }
  }

  uplink {
    name       = "storage"
    mode       = "bond"
    identifier = "mac_address"
    interfaces = ["00:1b:21:3c:4d:60", "00:1b:21:3c:4d:61"]
    lacp       = true

    host_network {
      subnet_id   = meltcloud_subnet.storage_ipam.id
      vlan_tagged = false
      primary     = false
    }
  }
}
