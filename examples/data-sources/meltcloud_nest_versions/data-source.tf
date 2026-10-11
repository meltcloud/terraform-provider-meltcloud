data "meltcloud_nest_versions" "all" {}

# the newest nest version that is not revoked
locals {
  newest_nest = [
    for version in data.meltcloud_nest_versions.all.versions : version.name if !version.revoked
  ][0]
}
