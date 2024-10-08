---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "meltcloud_uefi_http_boot_url Resource - meltcloud"
subcategory: ""
description: |-
  Generate UEFI HTTP Boot URLs https://meltcloud.io/docs/guides/boot-config/create-uefi-http-boot-urls.html for servers that support UEFI HTTP Boot.
---

# meltcloud_uefi_http_boot_url (Resource)

Generate [UEFI HTTP Boot URLs](https://meltcloud.io/docs/guides/boot-config/create-uefi-http-boot-urls.html) for servers that support UEFI HTTP Boot.

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

# create ipxe boot artifact
resource "meltcloud_ipxe_boot_artifact" "example" {
  name       = "my-artifact"
  expires_at = time_offset.in_a_year.rfc3339
}

# create UEFI HTTP Boot URL
resource "meltcloud_uefi_http_boot_url" "example" {
  ipxe_boot_artifact_id = meltcloud_ipxe_boot_artifact.example.id
  protocols             = "http_and_https"

  name       = "my-boot-url"
  expires_at = time_offset.in_a_year.rfc3339
}

# output url - can be used as DHCP option for servers that support UEFI HTTP Boot
output "uefi_http_boot_url" {
  value     = meltcloud_uefi_http_boot_url.example.http_url
  sensitive = true
}

# output url - can be used as DHCP option for servers that support UEFI HTTPS Boot
output "uefi_https_boot_url" {
  value     = meltcloud_uefi_http_boot_url.example.https_url
  sensitive = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `expires_at` (String) Timestamp when the URL should expire
- `ipxe_boot_artifact_id` (Number) Internal ID of the iPXE Boot Artifact that this UEFI HTTP Boot URL should be generated for
- `name` (String) Name of the UEFI HTTP Boot URL, not case-sensitive. Must be unique per iPXE Boot Artifact.
- `protocols` (String) Protocols to support. Must be either http_only, https_only or http_and_https.

### Read-Only

- `http_url_amd64` (String, Sensitive) HTTP URL of the UEFI HTTP Boot URL for the amd64 architecture. Is null if protocols is set to https_only.
- `http_url_arm64` (String, Sensitive) HTTP URL of the UEFI HTTP Boot URL for the arm64 architecture. Is null if protocols is set to https_only.
- `https_url_amd64` (String, Sensitive) HTTPS URL of the UEFI HTTP Boot URL for the amd64 architecture. Is null if protocols is set to http_only.
- `https_url_arm64` (String, Sensitive) HTTPS URL of the UEFI HTTP Boot URL for the arm64 architecture. Is null if protocols is set to http_only.
- `id` (Number) Internal ID of the UEFI HTTP Boot URL on meltcloud

## Import

Import is supported using the following syntax:

```shell
# Resource can be imported by using the resource path as displayed in the URL
terraform import meltcloud_uefi_http_boot_url.example ipxe_boot_artifacts/35/uefi_http_boot_urls/14
```
