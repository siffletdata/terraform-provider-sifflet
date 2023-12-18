terraform {
  required_providers {
    sifflet = {
      source = "hashicorp.com/edu/sifflet"
    }
  }
}

provider "sifflet" {
  host = "http://localhost:8000"
}

resource "sifflet_datasource" "test" {
  name      = "toto"
  secret_id = "abc"
  type      = "data_source"
  bigquery = {
    type               = "toto"
    billing_project_id = "bproject_id"
    dataset_id         = "dataset_id"
    project_id         = "project_id"
    timezone_data = {
      timezone   = "PAris"
      utc_offset = "+1"
    }
  }

}

# data "sifflet_datasources" "example" {}

# output "datasources" {
#   value = data.sifflet_datasources.example
# }
