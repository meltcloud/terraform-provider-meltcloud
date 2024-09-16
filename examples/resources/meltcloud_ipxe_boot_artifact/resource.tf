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