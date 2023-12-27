data "sifflet_datasources" "example" {}

output "datasources" {
  value = data.sifflet_datasources.example
}
