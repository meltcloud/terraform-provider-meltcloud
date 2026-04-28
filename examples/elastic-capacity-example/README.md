# Elastic Capacity (provider side)

This example represents the **provider** side of the elastic capacity flow. It
creates:

- A cluster to host the elastic capacity (`meltcloud_cluster`)
- A machine pool providing the underlying compute (`meltcloud_machine_pool`); joining a machine is left as an exercise to the reader (i.e. in the GUI)
- Cilium installed via Helm on the cluster
- An `meltcloud_elastic_capacity` backed by the cluster
- An `meltcloud_elastic_share` allocating a slice of that capacity to a
  consuming organization (identified by UUID)

The corresponding consumer-side example lives in
[`../elastic-pool-example`](../elastic-pool-example/) and creates an
`meltcloud_elastic_pool` that consumes the share produced here.

## Prerequisites

- A meltcloud API key for the provider organization
- The UUID of the organization that should provide the share
- The UUID of the organization that should consume the share

## Running

This is a two-phase apply because Cilium and the elastic capacity cannot come
up until at least one machine has joined the machine pool (otherwise cilium and kubevirt never come up).

**Phase 1** — create the cluster and machine pool:

```bash
export MELTCLOUD_API_KEY='<your-api-key>'
export TF_VAR_consuming_organization_uuid='<uuid-of-consumer-org>'

terraform init
terraform apply
```

Then, in the meltcloud GUI, assign a machine to `meltcloud_machine_pool.example`
and wait until it reports ready.

**Phase 2** — enable Cilium + the elastic capacity/share:

```bash
terraform apply -var deploy_capacity=true
```

Once applied, note the resulting `elastic_share_id` output — it is the input
for the [consumer-side example](./../elastic-pool-example/README.md).
