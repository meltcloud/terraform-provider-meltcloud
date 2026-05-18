# Elastic Node Pool (consumer side)

This example represents the **consumer** side of the elastic fleet flow. It
takes an existing `meltcloud_elastic_quota` (created by the provider) and
spins up a cluster backed by an `meltcloud_elastic_node_pool` that consumes that
quota. It creates:

- A cluster (`meltcloud_cluster`) with non-overlapping pod/service CIDRs
  relative to the outer provider cluster
- An `meltcloud_elastic_node_pool` consuming the given quota, with 1 node × 2
  cores × 1 GB RAM × 20 GB disk
- Cilium installed via Helm, with `tunnelPort=8474` and `MTU=1300` so its
  VXLAN traffic does not collide with the outer cluster's Cilium
- [podinfo](https://github.com/stefanprodan/podinfo) deployed as a sample
  workload (2 replicas) in the `default` namespace

The provider-side setup that produces the quota lives in
[`../elastic-fleet-example`](../elastic-fleet-example/).

## Prerequisites

- A meltcloud API key for the consuming organization
- The numeric ID of the `meltcloud_elastic_quota` to consume (produced by the
  provider-side example)
- The UUID of the organization for the elastic node pool ("consuming organization")

## Running

```bash
export MELTCLOUD_API_KEY='<your-api-key>'
export TF_VAR_consuming_organization_uuid='<uuid-of-consumer-org>'
export TF_VAR_elastic_quota_id=<id-of-quota>

terraform init
terraform apply
```
