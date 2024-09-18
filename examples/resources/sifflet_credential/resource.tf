resource "sifflet_credentials" "example" {
  name        = "credentialid"
  description = "Credential description."
  # Due to API limitations, Terraform can't detect changes to the value that are made outside of Terraform.
  value = "Credential value"
}
