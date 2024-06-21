terraform {
  required_providers {
    melt = {
      source = "meltcloud.io/melt/melt"
    }
  }
}

provider "melt" {
  endpoint     = "http://localhost:3000"
  organization = "deadbeef-0000-0000-0000-000000000000"
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

