# create cluster with auto-assigned network defaults
resource "meltcloud_cluster" "example" {
  name    = "melt02"
  version = "1.33"
}

# create cluster with explicit network configuration
resource "meltcloud_cluster" "example_custom_network" {
  name    = "melt03"
  version = "1.33"

  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

# create cluster whose control plane survives losing a Nest Instance
# (needs a scaled-out Nest and a highly available Foundry)
resource "meltcloud_cluster" "example_ha" {
  name              = "melt04"
  version           = "1.33"
  high_availability = true
}

# create cluster following the dev release channel
resource "meltcloud_cluster" "example_dev" {
  name            = "melt05"
  version         = "1.35"
  release_channel = "dev"
}

# create cluster pinned to a Kubernetes Bundle, version is derived from it
resource "meltcloud_cluster" "example_pinned" {
  name            = "melt06"
  release_channel = "manual"
  manual_version  = "1.35.5-melt.30"
}

# create cluster without kube-proxy, for a CNI that replaces it such as Cilium
# (changing it later adds or removes kube-proxy in place)
resource "meltcloud_cluster" "example_without_kube_proxy" {
  name             = "melt07"
  version          = "1.35"
  addon_kube_proxy = false
}

# use kubeconfig to install a helm chart, for example a CNI
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

resource "helm_release" "cilium" {
  name       = "cilium"
  repository = "https://helm.cilium.io"
  chart      = "cilium"
  namespace  = "kube-system"
  version    = "1.16.1"

  set {
    name  = "ipam.mode"
    value = "kubernetes"
  }
}