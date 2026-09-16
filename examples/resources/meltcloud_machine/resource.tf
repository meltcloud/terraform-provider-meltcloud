# create cluster
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.33"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

# create a machine pool
resource "meltcloud_machine_pool" "example" {
  cluster_id = meltcloud_cluster.example.id

  name    = "pool1"
  version = "1.32"
}

# register a machine assigned to the pool
resource "meltcloud_machine" "node1" {
  uuid = "0442228d-023e-42ab-af34-da267d3e9c37"
  name = "meltcloud-node01"

  network_profile_id       = meltcloud_network_profile.single_untagged.id
  depot_network_profile_id = meltcloud_network_profile.single_untagged.id

  machine_pool_id = meltcloud_machine_pool.example.id
}

# register an unassigned machine
resource "meltcloud_machine" "node2" {
  uuid = "8d8fd677-db06-4acf-ac34-920b950ddbe5"
  name = "meltcloud-node02"

  network_profile_id       = meltcloud_network_profile.single_untagged.id
  depot_network_profile_id = meltcloud_network_profile.auto.id

  label {
    key   = "topology.kubernetes.io/region"
    value = "ch"
  }

  label {
    key   = "topology.kubernetes.io/zone"
    value = "az1"
  }
}

# import a machine that self-registered during enrollment.
# The id can be found from the URL in foundry.
import {
  to = meltcloud_machine.worker1
  id = "machines/42"
}

resource "meltcloud_machine" "worker1" {
  uuid = "cb4b3a08-3c3d-4b0e-9a1e-2f1d9a2f7c11"
  name = "worker1"

  network_profile_id       = data.meltcloud_network_profile.default.id
  depot_network_profile_id = data.meltcloud_network_profile.default.id

  machine_pool_id = meltcloud_machine_pool.example.id
}

data "meltcloud_network_profile" "default" {
  name = "default"
}

