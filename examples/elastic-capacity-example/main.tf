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
  description = "UUID of the organization that provides the elastic share"
}

provider "meltcloud" {
  endpoint     = "https://app.192.168.101.101.d.meltcloud.io"
  organization = var.providing_organization_uuid
}

variable "consuming_organization_uuid" {
  type        = string
  description = "UUID of the organization that consumes the elastic share"
}

variable "deploy_capacity" {
  type        = bool
  default     = false
  description = "Whether to deploy Cilium and the elastic capacity/share. Only set to true after a machine has been assigned to the machine pool — before that, the cluster's API is unreachable and the capacity cannot come up."
}

# cluster that hosts the elastic capacity
resource "meltcloud_cluster" "example" {
  name             = "elastic-capacity-example"
  version          = "1.35"
  pod_cidr         = "10.38.0.0/16"
  service_cidr     = "10.98.0.0/16"
  dns_service_ip   = "10.98.0.10"
  addon_core_dns   = true
  addon_kube_proxy = true
}

# machine pool providing compute for the capacity
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
  count = var.deploy_capacity ? 1 : 0

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

# elastic capacity backed by the cluster
resource "meltcloud_elastic_capacity" "example" {
  count = var.deploy_capacity ? 1 : 0

  cluster_id = meltcloud_cluster.example.id
  name       = "capacity1"
}

# share of the capacity allocated to a consuming organization
resource "meltcloud_elastic_share" "example" {
  count = var.deploy_capacity ? 1 : 0

  capacity_id                 = meltcloud_elastic_capacity.example[0].id
  consuming_organization_uuid = var.consuming_organization_uuid

  name      = "share1"
  cores     = 100
  memory_mb = 102400
  disk_gb   = 1000
}

output "elastic_share_id" {
  description = "ID of the created elastic share — pass this to the consumer-side example as TF_VAR_elastic_share_id"
  value       = var.deploy_capacity ? meltcloud_elastic_share.example[0].id : null
}
