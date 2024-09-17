terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}

resource "random_uuid" "machine_override" {
}

data "meltcloud_ipxe_chain_url" "example_name" {
  name = "url1"
}

output "customized_ipxe_script" {
  sensitive = true
  value     = provider::meltcloud::customize_uuid_in_ipxe_script(data.meltcloud_ipxe_chain_url.example_name.script, random_uuid.machine_override.result)
}