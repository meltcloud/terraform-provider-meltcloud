# the nest installer creates the nest, import it
import {
  to = meltcloud_nest.this
  id = "nest"
}

resource "meltcloud_nest" "this" {
  version = "1.35.3-nest.67"
}
