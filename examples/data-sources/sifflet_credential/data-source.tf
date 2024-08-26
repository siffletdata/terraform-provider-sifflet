data "sifflet_credential" "example" {
  name = "credentialname"
}

output "credential_description" {
  value = data.sifflet_credential.example.description
}
