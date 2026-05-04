data "sifflet_domain" "example" {
  id = "4b0968b9-3a39-46fc-9480-cd117d8a0fbe"
}

resource "sifflet_user" "example" {
  email = "user@example.com"
  name  = "Example User"
  role  = "EDITOR"
  permissions = [{
    domain_id   = data.sifflet_domain.example.id
    domain_role = "VIEWER"
  }]
}

resource "sifflet_team" "example" {
  name        = "Data Engineering Team"
  description = "Team responsible for data pipelines and infrastructure"

  domain_permissions = [{
    domain_id   = data.sifflet_domain.example.id
    domain_role = "EDITOR"
  }]

  users = [{
    user_id = sifflet_user.example.id
  }]
}
