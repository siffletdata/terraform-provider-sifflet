data "sifflet_credentials" "example" {
  name = "credential-name"
}

output "credentials_description" {
  value = data.sifflet_credentials.example.description
}
