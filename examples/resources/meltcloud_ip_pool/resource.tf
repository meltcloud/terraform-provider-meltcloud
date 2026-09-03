resource "meltcloud_ip_pool" "example" {
  name        = "wl-prd"
  cidr        = "10.20.0.0/24"
  description = "workload addresses"

  # addresses are handed out from allocatable ranges only, so everything outside
  # one is left alone
  range {
    kind          = "allocatable"
    start_address = "10.20.0.10"
    end_address   = "10.20.0.200"
  }

  # an excluded range prevents handing out addresses something else already uses
  range {
    kind          = "excluded"
    start_address = "10.20.0.100"
    end_address   = "10.20.0.110"
    description   = "printers and switches"
  }
}
