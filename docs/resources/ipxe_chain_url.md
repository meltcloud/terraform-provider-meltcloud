---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "meltcloud_ipxe_chain_url Resource - meltcloud"
subcategory: ""
description: |-
  Generate iPXE Chain URLs https://meltcloud.io/docs/guides/boot-config/create-ipxe-chain-urls.html for providers that allow booting an iPXE Script or a remote iPXE URL (for example Equinix Metal)
---

# meltcloud_ipxe_chain_url (Resource)

Generate [iPXE Chain URLs](https://meltcloud.io/docs/guides/boot-config/create-ipxe-chain-urls.html) for providers that allow booting an iPXE Script or a remote iPXE URL (for example Equinix Metal)

## Example Usage

```terraform
terraform {
  required_providers {
    time = {
      source  = "hashicorp/time"
      version = "0.11.2"
    }
  }
}

# create expiry timestamp
resource "time_offset" "in_a_year" {
  offset_days = 365
}

# create ipxe chain url
resource "meltcloud_ipxe_boot_artifact" "example" {
  name       = "my-artifact"
  expires_at = time_offset.in_a_year.rfc3339
}

# download the iso
data "http" "ipxe_iso" {
  url = meltcloud_ipxe_boot_artifact.example.download_url_iso
}

# save iso to a file
resource "local_sensitive_file" "ipxe_iso" {
  filename        = "${path.module}/ipxe.iso"
  content_base64  = data.http.ipxe_iso.response_body_base64
  file_permission = "0600"
}

# output url - can be used for providers that support booting from remote URL
output "ipxe_chain_url" {
  value     = meltcloud_ipxe_chain_url.example.url
  sensitive = true
}

# output script - can be used for providers that support providing a full iPXE script
output "ipxe_chain_script" {
  value     = meltcloud_ipxe_chain_url.example.script
  sensitive = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `expires_at` (String) Timestamp when the URL should expire
- `name` (String) Name of the iPXE Chain URL

### Read-Only

- `id` (Number) Internal ID of the iPXE Chain URL on meltcloud
- `script` (String, Sensitive) The complete iPXE script
- `url` (String, Sensitive) URL to the iPXE chain script

## Import

Import is supported using the following syntax:

```shell
# Resource can be imported by specifying the numeric identifier.
terraform import meltcloud_ipxe_chain_url.example 123
```