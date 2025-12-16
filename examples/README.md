# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or ar testable even if some parts are not relevant for the documentation.

* **provider/provider.tf** example file for the provider index page
* **data-sources/`full data source name`/data-source.tf** example file for the named data source page
* **resources/`full resource name`/resource.tf** example file for the named data source page

## Running the Full Example

The `full-example` directory contains a complete working example that demonstrates the provider's resources and integrations. It's split into two separate Terraform configurations to handle dependencies properly:

1. **Main Infrastructure** (`main.tf`) - Creates the Meltcloud resources (cluster, machines, etc.)
2. **Helm Deployments** (`helm/main.tf`) - Deploys Helm charts using the cluster created in step 1

### Apply Steps

```bash
# 1. Apply meltcloud infrastructure
cd full-example
terraform init
terraform apply

# 2. Apply Helm configuration on top (reads cluster info from remote state)
cd helm
terraform init
terraform apply
```

The Helm configuration uses `terraform_remote_state` to automatically read the cluster kubeconfig from the main state file, ensuring the cluster exists before attempting to deploy charts.
