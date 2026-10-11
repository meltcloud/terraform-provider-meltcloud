# the nest installer creates the foundry, so it is imported
import {
  to = meltcloud_foundry.this
  id = "foundry"
}

resource "meltcloud_foundry" "this" {
  version = "1.5.0"

  # spread the foundry over nest instances in at least two zones
  ha_topology = "zone"

  identity_provider   = "auth0"
  auth0_domain        = "example.eu.auth0.com"
  auth0_client_id     = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  auth0_client_secret = var.auth0_client_secret

  letsencrypt_enabled = true
  letsencrypt_email   = "ops@example.com"
}

variable "auth0_client_secret" {
  type      = string
  sensitive = true
}
