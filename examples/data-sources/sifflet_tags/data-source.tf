data "sifflet_tags" "example" {}

output "tags" {
  value = data.sifflet_tags.example
}
