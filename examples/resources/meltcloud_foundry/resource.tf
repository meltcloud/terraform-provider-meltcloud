# the nest installer creates the foundry, import it
import {
  to = meltcloud_foundry.this
  id = "foundry"
}

resource "meltcloud_foundry" "this" {
  version = "1.5.0"

  identity_provider = "auth0"
  auth0_config = {
    domain                   = "example.eu.auth0.com"
    client_id                = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    client_secret_wo         = var.auth0_client_secret
    client_secret_wo_version = 1
  }

  letsencrypt_config = {
    enabled = true
    email   = "ops@example.com"
  }
}

variable "auth0_client_secret" {
  type      = string
  sensitive = true
}
