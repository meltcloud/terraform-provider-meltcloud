# get by ID
data "meltcloud_uefi_http_boot_url" "example_id" {
  ipxe_boot_artifact_id = 1
  id                    = 42
}

# get by name
data "meltcloud_uefi_http_boot_url" "example_name" {
  ipxe_boot_artifact_id = 1
  name                  = "url1"
}