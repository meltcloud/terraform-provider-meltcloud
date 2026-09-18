# get cluster by ID
data "meltcloud_cluster" "example_id" {
  id = 42
}

# get cluster by name
data "meltcloud_cluster" "example_name" {
  name = "melt01"
}