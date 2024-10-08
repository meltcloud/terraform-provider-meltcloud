---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "meltcloud_ipxe_boot_artifact Resource - meltcloud"
subcategory: ""
description: |-
  An iPXE Boot Artifact https://meltcloud.io/docs/guides/boot-config/create-ipxe-boot-artifacts.html contains a set of bootable images with an X509 client certificate to securely boot into your meltcloud organization.
---

# meltcloud_ipxe_boot_artifact (Resource)

An [iPXE Boot Artifact](https://meltcloud.io/docs/guides/boot-config/create-ipxe-boot-artifacts.html) contains a set of bootable images with an X509 client certificate to securely boot into your meltcloud organization.

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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `expires_at` (String) Timestamp when the artifact should expire
- `name` (String) Name of the iPXE Boot Artifact, not case-sensitive. Must be unique within the organization.

### Read-Only

- `download_url_efi_amd64` (String, Sensitive) URL to download the amd64 EFI boot artifact
- `download_url_efi_arm64` (String, Sensitive) URL to download the arm64 EFI boot artifact
- `download_url_iso` (String, Sensitive) URL to download the ISO
- `download_url_pxe` (String, Sensitive) URL to download the PCBIOS artifact (.undionly)
- `download_url_raw_amd64` (String, Sensitive) URL to download the amd64 Raw boot artifact
- `id` (Number) Internal ID of the iPXE Boot Artifact

## Import

Import is supported using the following syntax:

```shell
# Resource can be imported by using the resource path as displayed in the URL
terraform import meltcloud_ipxe_boot_artifact.example ipxe_boot_artifacts/35
```
