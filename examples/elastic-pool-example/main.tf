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

provider "meltcloud" {
  endpoint     = "https://app.192.168.101.101.d.meltcloud.io"
  organization = "fe97cd36-b6d9-4d8f-a89a-9b595bd00a8c"
}

variable "elastic_share_id" {
  type        = number
  description = "ID of the elastic share to consume for the pool"
}

data "meltcloud_elastic_share" "existing" {
  id = var.elastic_share_id
}

resource "meltcloud_cluster" "example" {
  name             = "elastic-pool-example"
  version          = "1.35"
  pod_cidr         = "10.37.0.0/16"
  service_cidr     = "10.97.0.0/16"
  dns_service_ip   = "10.97.0.10"
  addon_core_dns   = true
  addon_kube_proxy = true
}

resource "meltcloud_elastic_pool" "example" {
  cluster_id = meltcloud_cluster.example.id
  share_id   = data.meltcloud_elastic_share.existing.id

  name       = "pool1"
  version    = "1.35"
  node_count = 2

  node_config {
    cores     = 4
    memory_mb = 2048
    disk_gb   = 20
  }
}

# CNI for the inner cluster
provider "helm" {
  kubernetes = {
    host                   = meltcloud_cluster.example.kubeconfig.host
    client_certificate     = base64decode(meltcloud_cluster.example.kubeconfig.client_certificate)
    client_key             = base64decode(meltcloud_cluster.example.kubeconfig.client_key)
    cluster_ca_certificate = base64decode(meltcloud_cluster.example.kubeconfig.cluster_ca_certificate)
  }
}

resource "helm_release" "cilium" {
  name       = "cilium"
  repository = "https://helm.cilium.io"
  chart      = "cilium"
  namespace  = "kube-system"
  version    = "1.17.4"

  set = [
    {
      name  = "image.pullPolicy"
      value = "IfNotPresent"
    },
    {
      name  = "ipam.mode"
      value = "kubernetes"
    },
    # overrides VXLAN port and MTU so traffic does not
    # collide with the outer cluster's Cilium (default tunnelPort 8472, MTU 1500).
    {
      name  = "tunnelPort"
      value = "8474"
    },
    {
      name  = "MTU"
      value = "1300"
    },
  ]
}

resource "helm_release" "podinfo" {
  name       = "podinfo"
  repository = "oci://ghcr.io/stefanprodan/charts"
  chart      = "podinfo"
  namespace  = "default"

  set = [
    {
      name  = "replicaCount"
      value = "2"
    },
  ]

  depends_on = [helm_release.cilium]
}
