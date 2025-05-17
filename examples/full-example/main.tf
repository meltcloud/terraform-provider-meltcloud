terraform {
  required_providers {
    meltcloud = {
      source = "meltcloud/meltcloud"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.11.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}

provider "meltcloud" {
  endpoint     = "http://localhost:3000"
  organization = "deadbeef-0000-0000-0000-000000000000"
}

resource "time_offset" "in_a_year" {
  offset_days = 365
}

resource "meltcloud_enrollment_image" "example" {
  name                         = "my-image"
  expires_at                   = time_offset.in_a_year.rfc3339
  install_disk_device          = "/dev/vda"
  install_disk_force_overwrite = true
  vlan                         = 101
  enable_http                  = true
}

data "meltcloud_enrollment_image" "example_id" {
  id = meltcloud_enrollment_image.example.id
}

data "meltcloud_enrollment_image" "example_name" {
  name = meltcloud_enrollment_image.example.name
}

resource "meltcloud_ipxe_boot_artifact" "example" {
  name       = "my-artifact"
  expires_at = time_offset.in_a_year.rfc3339
}

data "meltcloud_ipxe_boot_artifact" "example_id" {
  id = meltcloud_ipxe_boot_artifact.example.id
}

data "meltcloud_ipxe_boot_artifact" "example_name" {
  name = meltcloud_ipxe_boot_artifact.example.name
}

resource "meltcloud_uefi_http_boot_url" "example" {
  ipxe_boot_artifact_id = meltcloud_ipxe_boot_artifact.example.id
  protocols             = "http_and_https"

  name       = "my-boot-url"
  expires_at = time_offset.in_a_year.base_rfc3339
}

data "meltcloud_uefi_http_boot_url" "example_id" {
  ipxe_boot_artifact_id = meltcloud_ipxe_boot_artifact.example.id
  id                    = meltcloud_uefi_http_boot_url.example.id
}

data "meltcloud_uefi_http_boot_url" "example_name" {
  ipxe_boot_artifact_id = meltcloud_ipxe_boot_artifact.example.id
  name                  = meltcloud_uefi_http_boot_url.example.name
}

resource "meltcloud_ipxe_chain_url" "example" {
  name       = "my-chain-url"
  expires_at = time_offset.in_a_year.rfc3339
}

data "meltcloud_ipxe_chain_url" "example_id" {
  id = meltcloud_ipxe_chain_url.example.id
}

data "meltcloud_ipxe_chain_url" "example_name" {
  name = meltcloud_ipxe_chain_url.example.name
}

resource "random_uuid" "machine_override" {
}

resource "meltcloud_cluster" "example" {
  name           = "melt03"
  version        = "1.30"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

data "meltcloud_cluster" "example_id" {
  id = meltcloud_cluster.example.id
}

data "meltcloud_cluster" "example_name" {
  name = meltcloud_cluster.example.name
}

resource "meltcloud_machine_pool" "example" {
  cluster_id = meltcloud_cluster.example.id

  name                = "pool1"
  version             = "1.29"
  primary_disk_device = "/dev/vda"

  network_profile_id = meltcloud_network_profile.example.id
}

resource "meltcloud_network_profile" "example" {
  name = "profile1"

  vlan {
    vlan      = 10
    dhcp      = true
    interface = "eth0"
  }

  vlan {
    vlan      = 2
    dhcp      = false
    interface = "eth1"
  }

  bridge {
    name      = "br0"
    interface = "br0"
    dhcp      = true
  }

  bridge {
    name      = "br1"
    interface = "br1"
    dhcp      = false
  }

  bond {
    name       = "james"
    kind       = "default"
    dhcp       = true
    interfaces = "eth4,eth5"
  }

  bond {
    name       = "other"
    kind       = "lacp"
    dhcp       = false
    interfaces = "eth6,eth7"
  }
}

data "meltcloud_machine_pool" "example_id" {
  cluster_id = meltcloud_cluster.example.id
  id         = meltcloud_machine_pool.example.id
}

resource "meltcloud_machine" "node1" {
  uuid = "0442228d-023e-42ab-af34-da267d3e9c37"
  name = "meltcloud-node01"

  machine_pool_id = meltcloud_machine_pool.example.id

  label {
    key   = "topology.kubernetes.io/region"
    value = "ch"
  }

  label {
    key   = "topology.kubernetes.io/zone"
    value = "az3"
  }
}

data "meltcloud_machine" "example_id" {
  id = meltcloud_machine.node1.id
}

data "meltcloud_machine" "example_uuid" {
  uuid = meltcloud_machine.node1.uuid
}

provider "helm" {
  kubernetes {
    host                   = meltcloud_cluster.example.kubeconfig.host
    username               = meltcloud_cluster.example.kubeconfig.username
    password               = meltcloud_cluster.example.kubeconfig.password
    client_certificate     = base64decode(meltcloud_cluster.example.kubeconfig.client_certificate)
    client_key             = base64decode(meltcloud_cluster.example.kubeconfig.client_key)
    cluster_ca_certificate = base64decode(meltcloud_cluster.example.kubeconfig.cluster_ca_certificate)
  }
}

resource "helm_release" "test" {
  name       = "example-chart"
  repository = "./"
  chart      = "example-chart"
}

