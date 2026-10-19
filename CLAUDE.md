# Agent Guide

- Remind the user to make sure PR and docs are well understandable from the perspective of a terraform consumer. Insist on user review of them when creating PRs.
- Test changes against a real foundry

## Schema

- Use nested attributes, never blocks: `SingleNestedAttribute` for one object, `ListNestedAttribute` or
  `SetNestedAttribute` for several. A set can't hold write-only attributes, and a sensitive attribute in a
  set hides the whole set in plans, so use a list there.
- Group the settings of one feature into a single nested attribute, with an explicit `enabled` flag when
  the feature can be off.
- Every resource gets a data source with the same attributes, without secrets. Look it up by `id` or
  `name` with `withLookupAttributes`.
- Secrets are write-only, never stored in the state: `<name>_wo` with a `<name>_wo_version` that sends the
  secret again when it changes. A `Sensitive` attribute is still stored, so it is no substitute.
  Write-only attributes need Terraform 1.11, and a computed nested attribute can't contain one.
- Descriptions use the wording of the docs (meltcloud/melt-docs) and the foundry UI, and link the docs page.
