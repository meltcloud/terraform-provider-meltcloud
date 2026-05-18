# get elastic node pool by ID
data "meltcloud_elastic_node_pool" "example" {
  cluster_id = 1
  id         = 42
}
