# existing quota provisioned by the fleet provider
data "meltcloud_elastic_quota" "existing" {
  id = 1
}

# create cluster
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.35"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

# create an elastic node pool consuming the quota
resource "meltcloud_elastic_node_pool" "example" {
  cluster_id       = meltcloud_cluster.example.id
  elastic_quota_id = data.meltcloud_elastic_quota.existing.id

  name       = "nodepool1"
  version    = "1.35"
  node_count = 1

  node_config {
    vcpus      = 4
    memory_mib = 2048
    disk_gib   = 20
  }
}
