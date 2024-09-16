terraform {
  required_providers {
    meltcloud = {
      source  = "meltcloud/meltcloud"
      version = "~> 1.0"
    }
  }
}

provider "meltcloud" {
  endpoint     = "https://app.meltcloud.io" # optional
  organization = "deadbeef-0000-0000-0000-000000000000"
  api_key      = "eyJf..." # better pass it via env var MELTCLOUD_API_KEY or a tfvars file
}

# Create a cluster
resource "meltcloud_cluster" "example" {
  # ...
}