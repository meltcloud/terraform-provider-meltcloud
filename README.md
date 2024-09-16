# terraform-provider-meltcloud

## Local development

Instruct terraform to use the local provider:

```bash 
cat > ~/.terraformrc <<EOF
provider_installation {

  dev_overrides {
      "meltcloud.io/melt/melt" = "path/to/go/bin"
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

cd examples/provider-install-verification
terraform init # only required once
terraform apply
```

or Run/Debug within Goland:
- Run/Debug `main.go` with program arguments `-debug`
- Export the variables printed on stdout before running `terraform apply`

