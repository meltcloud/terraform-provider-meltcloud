# one zone per data center in the region
resource "meltcloud_zone" "dc1" {
  region_id = meltcloud_region.example.id
  name      = "ch-west-1"
}

resource "meltcloud_zone" "dc2" {
  region_id = meltcloud_region.example.id
  name      = "ch-west-2"
}
