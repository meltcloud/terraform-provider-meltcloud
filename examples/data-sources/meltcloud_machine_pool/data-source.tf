# get machine pool by ID
data "meltcloud_machine_pool" "example_id" {
  cluster_id = 1
  id         = 42
}

# get machine pool by name
data "meltcloud_machine_pool" "example_name" {
  cluster_id = 1
  name       = "workers"
}
