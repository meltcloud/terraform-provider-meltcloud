# get elastic fleet by ID
data "meltcloud_elastic_fleet" "example_id" {
  id = 1
}

# get elastic fleet by name
data "meltcloud_elastic_fleet" "example_name" {
  name = "fleet01"
}
