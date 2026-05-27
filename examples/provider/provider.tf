terraform {
  required_providers {
    meltcloud = {
      source  = "meltcloud/meltcloud"
      version = "~> 1.0"
    }
  }
}

# Cloud-hosted meltcloud
provider "meltcloud" {
  endpoint     = "https://app.meltcloud.io" # optional
  organization = "deadbeef-0000-0000-0000-000000000000"
  api_key      = "eyJf..." # better pass it via env var MELTCLOUD_API_KEY or a tfvars file
}

# Self-hosted Foundry with a private CA certificate
provider "meltcloud" {
  alias        = "self_hosted"
  endpoint     = "https://app.foundry.example.com"
  organization = "deadbeef-0000-0000-0000-000000000000"
  api_key      = "eyJf..."

  ca_cert_file = "/path/to/foundry-ca.pem" # or use MELTCLOUD_CACERT env var
  # ca_cert_pem = "-----BEGIN CERTIFICATE-----\n..." # alternative: inline PEM
}

# Create a cluster
resource "meltcloud_cluster" "example" {
  # ...
}