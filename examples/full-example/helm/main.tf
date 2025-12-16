terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "< 3"
    }
  }
}

data "terraform_remote_state" "main" {
  backend = "local"

  config = {
    path = "../terraform.tfstate"
  }
}

provider "helm" {
  kubernetes {
    host                   = data.terraform_remote_state.main.outputs.cluster_kubeconfig.host
    username               = data.terraform_remote_state.main.outputs.cluster_kubeconfig.username
    password               = data.terraform_remote_state.main.outputs.cluster_kubeconfig.password
    client_certificate     = base64decode(data.terraform_remote_state.main.outputs.cluster_kubeconfig.client_certificate)
    client_key             = base64decode(data.terraform_remote_state.main.outputs.cluster_kubeconfig.client_key)
    cluster_ca_certificate = base64decode(data.terraform_remote_state.main.outputs.cluster_kubeconfig.cluster_ca_certificate)
  }
}

resource "helm_release" "test" {
  name       = "example-chart"
  repository = "../"
  chart      = "example-chart"
}
