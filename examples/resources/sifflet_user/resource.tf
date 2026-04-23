data "sifflet_domain" "example" {
  id = "4b0968b9-3a39-46fc-9480-cd117d8a0fbe"
}

resource "sifflet_user" "example" {
  email = "user@example.com"
  name  = "Example User"
  role  = "EDITOR"
  # VIEWER and EDITOR users must have at least one domain role.
  # ADMIN users have access to all domains, so none should be set.
  permissions = [{
    domain_id   = data.sifflet_domain.test.id
    domain_role = "VIEWER"
  }]
}
