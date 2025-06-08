# get image by ID
data "meltcloud_enrollment_image" "example_id" {
  id = 42
}

# get image by name
data "meltcloud_enrollment_image" "example_name" {
  name = "my-image"
}