terraform {
  required_providers {
    time = {
      source  = "hashicorp/time"
      version = "0.11.2"
    }
  }
}

# create expiry timestamp
resource "time_offset" "in_a_year" {
  offset_days = 365
}

# create ipxe chain url
resource "meltcloud_ipxe_chain_url" "example" {
  name       = "my-chain-url"
  expires_at = time_offset.in_a_year.rfc3339
}

# output url - can be used for providers that support booting from remote URL
output "ipxe_chain_url" {
  value     = meltcloud_ipxe_chain_url.example.url
  sensitive = true
}

# output script - can be used for providers that support providing a full iPXE script
output "ipxe_chain_script" {
  value     = meltcloud_ipxe_chain_url.example.script
  sensitive = true
}