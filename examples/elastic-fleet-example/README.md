# Elastic Fleet (provider side)

This example represents the **provider** side of the elastic fleet flow. It
creates:

- A cluster to host the elastic fleet (`meltcloud_cluster`)
- A machine pool providing the underlying compute (`meltcloud_machine_pool`); joining a machine is left as an exercise to the reader (i.e. in the GUI)
- Cilium installed via Helm on the cluster
- A `meltcloud_elastic_fleet` backed by the cluster
- A `meltcloud_elastic_quota` allocating a slice of that fleet to a
  consuming organization (identified by UUID)

The corresponding consumer-side example lives in
[`../elastic-node-pool-example`](../elastic-node-pool-example/) and creates an
`meltcloud_elastic_node_pool` that consumes the quota produced here.

## Prerequisites

- A meltcloud API key for the provider organization
- The UUID of the organization that should provide the quota
- The UUID of the organization that should consume the quota (may be the same as the provider org)

## Running

This is a two-phase apply because Cilium and the elastic fleet cannot come
up until at least one machine has joined the machine pool (otherwise cilium and
kubevirt never come up).

**Phase 1** — create the cluster and machine pool:

```bash
export MELTCLOUD_API_KEY='<your-api-key>'
export TF_VAR_providing_organization_uuid='<uuid-of-provider-org>'
export TF_VAR_consuming_organization_uuid='<uuid-of-consumer-org>'

terraform init
terraform apply
```

Then, in the meltcloud GUI, assign a machine to `meltcloud_machine_pool.example`
and wait until it reports ready.

**Phase 2** — enable Cilium + the elastic fleet/quota:

```bash
terraform apply -var deploy_fleet=true
```

Once applied, note the resulting `elastic_quota_id` output — it is the input
for the [consumer-side example](./../elastic-node-pool-example/README.md).
