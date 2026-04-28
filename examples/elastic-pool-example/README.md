# Elastic Pool (consumer side)

This example represents the **consumer** side of the elastic capacity flow. It
takes an existing `meltcloud_elastic_share` (created by the provider) and
spins up a cluster backed by an `meltcloud_elastic_pool` that consumes that
share. It creates:

- A cluster (`meltcloud_cluster`) with non-overlapping pod/service CIDRs
  relative to the outer provider cluster
- An `meltcloud_elastic_pool` consuming the given share, with 2 nodes × 4
  cores × 2 GB RAM × 20 GB disk
- Cilium installed via Helm, with `tunnelPort=8474` and `MTU=1300` so its
  VXLAN traffic does not collide with the outer cluster's Cilium
- [podinfo](https://github.com/stefanprodan/podinfo) deployed as a sample
  workload (2 replicas) in the `default` namespace

The provider-side setup that produces the share lives in
[`../elastic-capacity-example`](../elastic-capacity-example/).

## Prerequisites

- A meltcloud API key for the consuming organization
- The numeric ID of the `meltcloud_elastic_share` to consume (produced by the
  provider-side example)
- The UUID of the organization for the elastic pool ("consuming organization")

## Running

```bash
export MELTCLOUD_API_KEY='<your-api-key>'
export TF_VAR_consuming_organization_uuid='<uuid-of-consumer-org>'
export TF_VAR_elastic_share_id=<id-of-share>

terraform init
terraform apply
```
