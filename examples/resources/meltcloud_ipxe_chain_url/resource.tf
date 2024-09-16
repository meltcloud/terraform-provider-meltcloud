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
resource "meltcloud_ipxe_boot_artifact" "example" {
  name       = "my-artifact"
  expires_at = time_offset.in_a_year.rfc3339
}

# download the iso
data "http" "ipxe_iso" {
  url = meltcloud_ipxe_boot_artifact.example.download_url_iso
}

# save iso to a file
resource "local_sensitive_file" "ipxe_iso" {
  filename        = "${path.module}/ipxe.iso"
  content_base64  = data.http.ipxe_iso.response_body_base64
  file_permission = "0600"
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