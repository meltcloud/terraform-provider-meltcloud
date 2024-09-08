terraform {
  required_providers {
    melt = {
      source = "meltcloud.io/melt/melt"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.11.2"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "2.15.0"
    }
  }
}

provider "melt" {
  endpoint     = "http://localhost:3000"
  organization = "deadbeef-0000-0000-0000-000000000000"
  api_key      = "eyJfcmFpbHMiOnsiZGF0YSI6WzEwXSwicHVyIjoiQXBpS2V5XG5hY2Nlc3NcbiJ9fQ==--d9d1bdddff5e8b1aee160e03a0e431801664e998"
}

resource "melt_cluster" "example" {
  name           = "melt02"
  version        = "1.29"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

resource "melt_machine_pool" "example" {
  cluster_id = melt_cluster.example.id

  name                = "pool2"
  version             = "1.29"
  primary_disk_device = "/dev/vda"
}

resource "melt_machine" "example" {
  #machine_pool_id = melt_machine_pool.example.id

  uuid = "2005cc24-522a-4485-9b9a-e60a61d9f9cf"
  name = "melt-node02"
}

resource "time_offset" "in_a_year" {
  offset_days = 365
}

resource "melt_ipxe_boot_artifact" "example" {
  name       = "tf-test2"
  expires_at = time_offset.in_a_year.rfc3339
}

# data "http" "ipxe_iso" {
#   url = melt_ipxe_boot_artifact.example.download_url_iso
# }
#
# resource "local_sensitive_file" "ipxe_iso" {
#   filename        = "${path.module}/ipxe.iso"
#   content_base64  = data.http.ipxe_iso.response_body_base64
#   file_permission = "0600"
# }

resource "melt_ipxe_chain_url" "example" {
  name       = "example"
  expires_at = time_offset.in_a_year.rfc3339
}

output "ipxe_chain_script" {
  value     = melt_ipxe_chain_url.example.script
  sensitive = true
}

# provider "helm" {
#   kubernetes {
#     host     = melt_cluster.example.kubeconfig.host
#     username = melt_cluster.example.kubeconfig.username
#     password = melt_cluster.example.kubeconfig.password
#     client_certificate = base64decode(melt_cluster.example.kubeconfig.client_certificate)
#     client_key = base64decode(melt_cluster.example.kubeconfig.client_key)
#     cluster_ca_certificate = base64decode(melt_cluster.example.kubeconfig.cluster_ca_certificate)
#   }
# }
#
# resource "helm_release" "cilium" {
#   name       = "cilium"
#   repository = "https://helm.cilium.io"
#   chart      = "cilium"
#   namespace  = "kube-system"
#   version    = "1.16.1"
#
#   set {
#     name  = "ipam.mode"
#     value = "kubernetes"
#   }
# }