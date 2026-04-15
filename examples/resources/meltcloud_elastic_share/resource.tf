# create cluster and capacity
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.35"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

resource "meltcloud_elastic_capacity" "example" {
  cluster_id = meltcloud_cluster.example.id
  name       = "capacity1"
}

# allocate a share of the capacity to a consuming organization
resource "meltcloud_elastic_share" "example" {
  capacity_id                 = meltcloud_elastic_capacity.example.id
  consuming_organization_uuid = "deadbeef-0000-0000-0000-000000000000"

  name      = "share1"
  cores     = 100
  memory_mb = 102400
  disk_gb   = 1000
}
