# get elastic node pool by ID
data "meltcloud_elastic_node_pool" "example_id" {
  cluster_id = 1
  id         = 42
}

# get elastic node pool by name
data "meltcloud_elastic_node_pool" "example_name" {
  cluster_id = 1
  name       = "workers"
}
