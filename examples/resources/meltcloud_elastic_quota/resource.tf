# create cluster and fleet
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.35"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

resource "meltcloud_elastic_fleet" "example" {
  cluster_id = meltcloud_cluster.example.id
  name       = "fleet1"
}

# allocate a quota of the fleet to a consuming organization
resource "meltcloud_elastic_quota" "example" {
  elastic_fleet_id            = meltcloud_elastic_fleet.example.id
  consuming_organization_uuid = "deadbeef-0000-0000-0000-000000000000"

  name       = "quota1"
  vcpus      = 100
  memory_mib = 102400
  disk_gib   = 1000
}
