# terraform-provider-meltcloud

## Quickstarts
- [Quickstart with Terraform | Docs | meltcloud.io](https://meltcloud.io/docs/guides/quick-start-terraform.html)
- [Documentation on Terraform Registry](https://registry.terraform.io/providers/meltcloud/meltcloud/latest/docs)


## Local development

Instruct terraform to use the local provider:

```bash 
cat > ~/.terraformrc <<EOF
provider_installation {

  dev_overrides {
      "registry.terraform.io/meltcloud/meltcloud" = "path/to/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
terraform-provider
EOF
```

Run from CLI:
``` 
go install

cd examples
terraform init # only required once

export MELTCLOUD_API_TOKEN=...
terraform apply
```

or Run/Debug within Goland:
- Run/Debug `main.go` with program arguments `-debug` and environment variables `MELTCLOUD_API_TOKEN=...`
- Export the variables printed on stdout before running `terraform apply`

