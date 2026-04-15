# create cluster
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.35"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

# create elastic capacity backed by the cluster
resource "meltcloud_elastic_capacity" "example" {
  cluster_id = meltcloud_cluster.example.id
  name       = "capacity1"
}
