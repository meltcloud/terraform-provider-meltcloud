terraform {
  required_providers {
    equinix = {
      source  = "equinix/equinix"
      version = "2.4.1"
    }
    meltcloud = {
      source  = "meltcloud/meltcloud"
      version = "~> 1.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.11.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
    local = {
      source  = "hashicorp/local"
      version = "2.5.2"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "2.15.0"
    }
  }
}

# initialize providers.
# set your meltcloud API Key via 'export MELTCLOUD_API_KEY=eyJ..' before terraform apply
provider "meltcloud" {
  # adapt to your organization
  organization = "f505052b-19cf-4761-b9ac-482fb3481297"
}

# set your API Key via 'export MELTCLOUD_API_KEY=eyJ..' before terraform apply
# set your equinix API key via 'export METAL_AUTH_TOKEN=zEq..' before terraform apply
provider "equinix" {
}

# create a cluster on meltcloud
resource "meltcloud_cluster" "equinix" {
  name           = "melt-equinix"
  version        = "1.30"
  pod_cidr       = "10.36.0.0/16"
  service_cidr   = "10.96.0.0/16"
  dns_service_ip = "10.96.0.10"
}

# save the kubeconfig for use with kubectl
resource "local_sensitive_file" "kubeconfig" {
  filename        = "${path.module}/melt-equinix.kubeconfig"
  content         = meltcloud_cluster.equinix.kubeconfig_raw
  file_permission = "0600"
}

# create a machine pool for equinix workers
resource "meltcloud_machine_pool" "equinix" {
  cluster_id = meltcloud_cluster.equinix.id

  name    = "equinix-pool"
  version = "1.30"
}

# pre-register the machine on meltcloud and assign it to the pool
resource "meltcloud_machine" "equinix01" {
  machine_pool_id = meltcloud_machine_pool.equinix.id

  uuid = equinix_metal_device.equinix01.id
  name = "melt-equinix-01"

  // TODO this currently won't work unless we implement update_existing = true
}

# create enrollment image for boot
resource "time_offset" "in_a_day" {
  offset_days = 1
}

resource "meltcloud_enrollment_image" "equinix" {
  name       = "equinix"
  expires_at = time_offset.in_a_day.rfc3339

  # equinix has the ephemeral disk on /dev/sda
  install_disk_device = "/dev/sda"
}

# create a bare metal machine!
resource "equinix_metal_device" "equinix01" {
  hostname         = "melt-equinix-01"
  plan             = "c3.small.x86"
  metro            = "fr"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"

  # adapt to your project
  project_id = "f46295d8-833a-4b96-a5e5-8e85ce2d471d"
  always_pxe = "true"
  user_data  = meltcloud_enrollment_image.equinix.ipxe_script_https_amd64
}

provider "helm" {
  kubernetes {
    host                   = meltcloud_cluster.equinix.kubeconfig.host
    username               = meltcloud_cluster.equinix.kubeconfig.username
    password               = meltcloud_cluster.equinix.kubeconfig.password
    client_certificate     = base64decode(meltcloud_cluster.equinix.kubeconfig.client_certificate)
    client_key             = base64decode(meltcloud_cluster.equinix.kubeconfig.client_key)
    cluster_ca_certificate = base64decode(meltcloud_cluster.equinix.kubeconfig.cluster_ca_certificate)
  }
}

# install a CNI so that the Kubernetes cluster/nodes becomes ready
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
