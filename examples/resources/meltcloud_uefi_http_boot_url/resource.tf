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

# create ipxe boot artifact
resource "meltcloud_ipxe_boot_artifact" "example" {
  name       = "my-artifact"
  expires_at = time_offset.in_a_year.rfc3339
}

# create UEFI HTTP Boot URL
resource "meltcloud_uefi_http_boot_url" "example" {
  ipxe_boot_artifact_id = meltcloud_ipxe_boot_artifact.example.id
  protocols             = "http_and_https"

  name       = "my-boot-url"
  expires_at = time_offset.in_a_year.rfc3339
}

# output url - can be used as DHCP option for servers that support UEFI HTTP Boot
output "uefi_http_boot_url" {
  value     = meltcloud_uefi_http_boot_url.example.http_url
  sensitive = true
}

# output url - can be used as DHCP option for servers that support UEFI HTTPS Boot
output "uefi_https_boot_url" {
  value     = meltcloud_uefi_http_boot_url.example.https_url
  sensitive = true
}