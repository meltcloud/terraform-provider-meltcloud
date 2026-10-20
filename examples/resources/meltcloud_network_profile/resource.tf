# single interface, auto-detected
resource "meltcloud_network_profile" "auto" {
  name = "workload"

  uplinks = [
    {
      name = "workload"
      mode = "auto"

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
      ]
    },
  ]
}

# single interface named, untagged vlan
resource "meltcloud_network_profile" "single_untagged" {
  name = "workload-single"

  uplinks = [
    {
      name       = "workload"
      mode       = "single"
      identifier = "kernel_name"
      interfaces = ["eth0"]

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
      ]
    },
  ]
}

# single interface named, tagged vlan
resource "meltcloud_network_profile" "single_tagged" {
  name = "workload-tagged"

  uplinks = [
    {
      name       = "workload"
      mode       = "single"
      identifier = "kernel_name"
      interfaces = ["eth0"]

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.storage_dhcp.id
          vlan_tagged = true
          primary     = true
        },
      ]
    },
  ]
}

# two interfaces bonded, untagged vlan
resource "meltcloud_network_profile" "bonded_untagged" {
  name = "workload-bonded"

  uplinks = [
    {
      name       = "workload"
      mode       = "bond"
      identifier = "kernel_name"
      interfaces = ["eth0", "eth1"]
      lacp       = true

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
      ]
    },
  ]
}

# two interfaces bonded, tagged vlan
resource "meltcloud_network_profile" "bonded_tagged" {
  name = "storage-bonded"

  uplinks = [
    {
      name       = "storage"
      mode       = "bond"
      identifier = "kernel_name"
      interfaces = ["eth0", "eth1"]
      lacp       = true

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.storage_dhcp.id
          vlan_tagged = true
          primary     = true
        },
      ]
    },
  ]
}

# single interface named, multiple subnets, all tagged
resource "meltcloud_network_profile" "single_multiple_tagged" {
  name = "main-single"

  uplinks = [
    {
      name       = "main"
      mode       = "single"
      identifier = "kernel_name"
      interfaces = ["eth0"]

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = true
          primary     = true
        },
        {
          subnet_id   = meltcloud_subnet.storage_dhcp.id
          vlan_tagged = true
          primary     = false
        },
      ]
    },
  ]
}

# single interface named, multiple subnets, one native and tagged subnets
resource "meltcloud_network_profile" "single_native_and_tagged" {
  name = "main-single-native"

  uplinks = [
    {
      name       = "main"
      mode       = "single"
      identifier = "kernel_name"
      interfaces = ["eth0"]

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
        {
          subnet_id   = meltcloud_subnet.storage_dhcp.id
          vlan_tagged = true
          primary     = false
        },
      ]
    },
  ]
}

# one bond, multiple subnets, all tagged
resource "meltcloud_network_profile" "bonded_multiple_tagged" {
  name = "main-bonded-tagged"

  uplinks = [
    {
      name       = "main"
      mode       = "bond"
      identifier = "kernel_name"
      interfaces = ["eth0", "eth1"]
      lacp       = true

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = true
          primary     = true
        },
        {
          subnet_id   = meltcloud_subnet.storage_dhcp.id
          vlan_tagged = true
          primary     = false
        },
      ]
    },
  ]
}

# one bond, multiple subnets, one native and tagged subnets
resource "meltcloud_network_profile" "bonded_native_and_tagged" {
  name = "main"

  uplinks = [
    {
      name       = "main"
      mode       = "bond"
      identifier = "kernel_name"
      interfaces = ["eth0", "eth1"]
      lacp       = true

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
        {
          subnet_id   = meltcloud_subnet.storage_dhcp.id
          vlan_tagged = true
          primary     = false
        },
      ]
    },
  ]
}

# multiple uplinks with a single interface each
resource "meltcloud_network_profile" "two_uplinks" {
  name = "workload-storage"

  uplinks = [
    {
      name       = "workload"
      mode       = "single"
      identifier = "kernel_name"
      interfaces = ["eth0"]

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
      ]
    },

    {
      name       = "storage"
      mode       = "single"
      identifier = "kernel_name"
      interfaces = ["eth1"]

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.storage_ipam.id
          vlan_tagged = false
          primary     = false
        },
      ]
    },
  ]
}

# multiple uplinks with bonds each
resource "meltcloud_network_profile" "two_uplinks_bonded" {
  name = "workload-storage-bonded"

  uplinks = [
    {
      name       = "workload"
      mode       = "bond"
      identifier = "mac_address"
      interfaces = ["00:1b:21:3c:4d:5e", "00:1b:21:3c:4d:5f"]
      lacp       = true

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.workload_dhcp.id
          vlan_tagged = false
          primary     = true
        },
      ]
    },

    {
      name       = "storage"
      mode       = "bond"
      identifier = "mac_address"
      interfaces = ["00:1b:21:3c:4d:60", "00:1b:21:3c:4d:61"]
      lacp       = true

      host_networks = [
        {
          subnet_id   = meltcloud_subnet.storage_ipam.id
          vlan_tagged = false
          primary     = false
        },
      ]
    },
  ]
}
