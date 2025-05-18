# create cluster
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.30"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

# create a machine pool
resource "meltcloud_machine_pool" "example" {
  cluster_id = meltcloud_cluster.example.id

  name    = "pool1"
  version = "1.29"
}

# register a machine assigned to the pool
resource "meltcloud_machine" "node1" {
  uuid = "0442228d-023e-42ab-af34-da267d3e9c37"
  name = "meltcloud-node01"

  machine_pool_id = meltcloud_machine_pool.example.id
}

# register an unassigned machine
resource "meltcloud_machine" "node2" {
  uuid = "8d8fd677-db06-4acf-ac34-920b950ddbe5"
  name = "meltcloud-node02"

  label {
    key   = "topology.kubernetes.io/region"
    value = "ch"
  }

  label {
    key   = "topology.kubernetes.io/zone"
    value = "az1"
  }
}