terraform {
  required_providers {
    meltcloud = {
      source = "meltcloud/meltcloud"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
  }
}

variable "providing_organization_uuid" {
  type        = string
  description = "UUID of the organization that provides the elastic quota"
}

provider "meltcloud" {
  endpoint     = "https://app.192.168.101.101.d.meltcloud.io"
  organization = var.providing_organization_uuid
}

variable "consuming_organization_uuid" {
  type        = string
  description = "UUID of the organization that consumes the elastic quota"
}

variable "deploy_fleet" {
  type        = bool
  default     = false
  description = "Whether to deploy Cilium and the elastic fleet/quota. Only set to true after a machine has been assigned to the machine pool — before that, the cluster's API is unreachable and the fleet cannot come up."
}

# cluster that hosts the elastic fleet
resource "meltcloud_cluster" "example" {
  name             = "elastic-fleet-example"
  version          = "1.35"
  pod_cidr         = "10.38.0.0/16"
  service_cidr     = "10.98.0.0/16"
  dns_service_ip   = "10.98.0.10"
  addon_core_dns   = true
  addon_kube_proxy = true
}

# machine pool providing compute for the fleet
resource "meltcloud_machine_pool" "example" {
  cluster_id = meltcloud_cluster.example.id

  name    = "pool1"
  version = "1.35"
}


provider "helm" {
  kubernetes = {
    host                   = meltcloud_cluster.example.kubeconfig.host
    client_certificate     = base64decode(meltcloud_cluster.example.kubeconfig.client_certificate)
    client_key             = base64decode(meltcloud_cluster.example.kubeconfig.client_key)
    cluster_ca_certificate = base64decode(meltcloud_cluster.example.kubeconfig.cluster_ca_certificate)
  }
}

# CNI for the outer cluster
resource "helm_release" "cilium" {
  count = var.deploy_fleet ? 1 : 0

  name       = "cilium"
  repository = "https://helm.cilium.io"
  chart      = "cilium"
  namespace  = "kube-system"
  version    = "1.17.4"

  set = [
    {
      name  = "ipam.mode"
      value = "kubernetes"
    },
  ]
}

# elastic fleet backed by the cluster
resource "meltcloud_elastic_fleet" "example" {
  count = var.deploy_fleet ? 1 : 0

  cluster_id = meltcloud_cluster.example.id
  name       = "fleet1"
}

# quota of the fleet allocated to a consuming organization
resource "meltcloud_elastic_quota" "example" {
  count = var.deploy_fleet ? 1 : 0

  elastic_fleet_id            = meltcloud_elastic_fleet.example[0].id
  consuming_organization_uuid = var.consuming_organization_uuid

  name       = "quota1"
  vcpus      = 100
  memory_mib = 102400
  disk_gib   = 1000
}

output "elastic_quota_id" {
  description = "ID of the created elastic quota — pass this to the consumer-side example as TF_VAR_elastic_quota_id"
  value       = var.deploy_fleet ? meltcloud_elastic_quota.example[0].id : null
}
