# create cluster
resource "meltcloud_cluster" "example" {
  name           = "melt02"
  version        = "1.30"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
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