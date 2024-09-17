# get machine by ID
data "meltcloud_machine_pool" "example_id" {
  cluster_id = 1
  id         = 42
}