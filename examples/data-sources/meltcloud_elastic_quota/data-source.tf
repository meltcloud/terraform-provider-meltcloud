# get elastic quota by ID
data "meltcloud_elastic_quota" "example_id" {
  id = 1
}

# get elastic quota by name, which is unique within its fleet
data "meltcloud_elastic_quota" "example_name" {
  elastic_fleet_id = 1
  name             = "gold"
}
