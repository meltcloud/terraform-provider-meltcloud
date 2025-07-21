# terraform-provider-meltcloud

## Quickstarts

- [Documentation on Terraform Registry](https://registry.terraform.io/providers/meltcloud/meltcloud/latest/docs)

## Local development

Instruct terraform to use the local provider:

```bash
cat > ~/.terraformrc <<EOF
provider_installation {

  dev_overrides {
    "registry.terraform.io/meltcloud/meltcloud" = "path/to/go/bin" # E.g. "/home/<user>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
EOF
```

Run from CLI:

```bash
go install

cd examples/full-example
terraform init # only required once

export MELTCLOUD_API_KEY=...
terraform apply
```

or Run/Debug within Goland:

- Run/Debug `main.go` with program arguments `-debug` and environment variables `MELTCLOUD_API_TOKEN=...`
- Export the variables printed on stdout before running `terraform apply`

## Releasing

- Generate the docs: `go generate`