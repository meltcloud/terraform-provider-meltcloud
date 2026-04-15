# existing share provisioned by the capacity provider
data "meltcloud_elastic_share" "existing" {
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

# create an elastic pool consuming the share
resource "meltcloud_elastic_pool" "example" {
  cluster_id = meltcloud_cluster.example.id
  share_id   = data.meltcloud_elastic_share.existing.id

  name       = "pool1"
  version    = "1.35"
  node_count = 1

  node_config {
    cores     = 4
    memory_mb = 2048
    disk_gb   = 20
  }
}
