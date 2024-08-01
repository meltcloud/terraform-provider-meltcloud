terraform {
  required_providers {
    melt = {
      source = "meltcloud.io/melt/melt"
    }
    time = {
      source = "hashicorp/time"
      version = "0.11.2"
    }
  }
}

provider "melt" {
  endpoint     = "http://localhost:3000"
  organization = "deadbeef-0000-0000-0000-000000000000"
  api_key      = "eyJfcmFpbHMiOnsiZGF0YSI6WzE5XSwicHVyIjoiQXBpS2V5XG5hY2Nlc3NcbiJ9fQ==--38d4cc47ac4e79328f918c75daed5cd248173d17"
}

resource "melt_cluster" "example" {
  name    = "melt02"
  version = "1.28"
}

resource "melt_machine_pool" "example" {
  cluster_id = melt_cluster.example.id

  name                = "pool2"
  version             = "1.28"
  primary_disk_device = "/dev/vda"
}

resource "melt_machine" "example" {
  machine_pool_id = melt_machine_pool.example.id

  uuid = "2005cc24-522a-4485-9b9a-e60a61d9f9cf"
  name = "melt-node01"
}

resource "time_offset" "in_a_year" {
  offset_days = 365
}

resource "melt_ipxe_boot_artifact" "example" {
  name = "tf-test2"
  expires_at = time_offset.in_a_year.rfc3339
}

data "http" "ipxe_iso" {
  url = melt_ipxe_boot_artifact.example.download_url_iso
}

resource "local_sensitive_file" "ipxe_iso" {
  filename        = "${path.module}/ipxe.iso"
  content_base64  = data.http.ipxe_iso.response_body_base64
  file_permission = "0600"
}

resource "melt_ipxe_chain_url" "example" {
  name       = "example"
  expires_at = time_offset.in_a_year.rfc3339
}

output "ipxe_chain_script" {
  value = melt_ipxe_chain_url.example.script
  sensitive = true
}

