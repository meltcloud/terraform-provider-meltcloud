data "meltcloud_foundry_versions" "all" {}

# the newest foundry version on the stable channel
locals {
  newest_stable_foundry = [
    for version in data.meltcloud_foundry_versions.all.versions : version.name
    if contains(version.channels, "stable") && !version.revoked
  ][0]
}
