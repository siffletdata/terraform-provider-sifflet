data "sifflet_source_v2_schemas" "example" {
  source_id = "00000000-0000-0000-0000-000000000000"
}

output "schemas" {
  value = data.sifflet_source_v2_schemas.example.schemas
}

output "first_schema_uri" {
  value = data.sifflet_source_v2_schemas.example.schemas[0].uri
}
