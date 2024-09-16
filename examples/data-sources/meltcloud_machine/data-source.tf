# get machine by ID
data "meltcloud_machine" "example_id" {
  id = 42
}

# get machine by UUID
data "meltcloud_machine" "example_uuid" {
  uuid = "0442228d-023e-42ab-af34-da267d3e9c37"
}