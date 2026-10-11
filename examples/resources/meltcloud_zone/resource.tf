# one zone per data center in the region
resource "meltcloud_zone" "zrh_dc1" {
  region_id = meltcloud_region.zurich.id
  name      = "zrh-dc1"
}

resource "meltcloud_zone" "zrh_dc2" {
  region_id = meltcloud_region.zurich.id
  name      = "zrh-dc2"
}
