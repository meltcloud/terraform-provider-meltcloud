# get zone by name within its region
data "meltcloud_zone" "example" {
  region_id = data.meltcloud_region.example.id
  name      = "ch-west-1"
}
