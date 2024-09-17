# get by ID
data "meltcloud_ipxe_chain_url" "example_id" {
  id = 42
}

# get by name
data "meltcloud_ipxe_chain_url" "example_name" {
  name = "url1"
}