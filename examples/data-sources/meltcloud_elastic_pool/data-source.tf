# get elastic pool by ID
data "meltcloud_elastic_pool" "example" {
  cluster_id = 1
  id         = 42
}
