resource "meltcloud_ip_pool" "example" {
  name        = "workload"
  cidr        = "10.20.0.0/24"
  description = "workload addresses"

  # addresses are handed out from allocatable ranges only
  range {
    kind          = "allocatable"
    start_address = "10.20.0.10"
    end_address   = "10.20.0.200"
  }

  # an excluded range prevents handing out addresses something else already uses
  range {
    kind          = "excluded"
    start_address = "10.20.0.1"
    end_address   = "10.20.0.3"
    description   = "routers"
  }
}

resource "meltcloud_ip_pool" "storage" {
  name = "storage"
  cidr = "10.30.0.0/24"

  range {
    kind          = "allocatable"
    start_address = "10.30.0.10"
    end_address   = "10.30.0.200"
  }
}
