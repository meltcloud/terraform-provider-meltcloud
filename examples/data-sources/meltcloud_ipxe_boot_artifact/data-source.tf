# get by ID
data "meltcloud_ipxe_boot_artifact" "example_id" {
  id = 42
}

# get by name
data "meltcloud_ipxe_boot_artifact" "example_name" {
  name = "artifact1"
}