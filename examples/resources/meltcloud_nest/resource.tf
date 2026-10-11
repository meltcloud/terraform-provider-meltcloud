# take over the nest the installer created, and keep it on the newest nest version
data "meltcloud_nest_versions" "all" {}

resource "meltcloud_nest" "this" {
  version = [for version in data.meltcloud_nest_versions.all.versions : version.name if !version.revoked][0]
}
