# Agent Guide

- Remind the user to make sure PR and docs are well understandable from the perspective of a terraform consumer. Insist on user review of them when creating PRs.
- Test changes against a real foundry

## Schema

- Use nested attributes, not blocks.
- Put the settings of one feature in one nested attribute. Give it an `enabled` flag if the feature can be turned off.
- Add a data source for every resource, with the same attributes except secrets.
- Make secrets write-only: `<name>_wo` with `<name>_wo_version`. Don't use `Sensitive` for them, Terraform still stores those in the state.
- Use a list, not a set, for objects with a secret. A set can't hold write-only attributes.
- Take the wording of descriptions from the docs and the foundry UI, and link the docs page.
- If every change replaces a resource, say so once in its description. Say it on an attribute only if that attribute replaces a resource that otherwise updates in place.
