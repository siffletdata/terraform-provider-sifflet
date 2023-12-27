terraform {
  required_providers {
    sifflet = {
      source = "hashicorp.com/edu/sifflet"
    }
  }
}

provider "sifflet" {
  host = "http://localhost:8080"
}

resource "sifflet_datasource" "test" {
  name            = "toto"
  secret_id       = "projects/369319181553/secrets/sifflet_servier_dev_bigquery_sa"
  cron_expression = "@daily"
  bigquery = {
    billing_project_id = "bproject_id"
    dataset_id         = "dataset_id"
    project_id         = "project_id"
    # timezone_data = {
    #   timezone   = "PAris"
    #   utc_offset = "+1"
    # }
  }

  # dbt = {
  #   project_name = "dbt_project_name"
  #   target       = "dbt_target"
  #   timezone_data = {
  #     timezone   = "PAris"
  #     utc_offset = "(UTC+01:00)"
  #   }
  # }

}


# resource "sifflet_datasource" "snowflake" {
#   name            = "snwoflake-orders"
#   secret_id       = "projects/369319181553/secrets/sifflet_servier_dev_bigquery_sa"
#   cron_expression = "@daily"
#   snowflake = {
#     account_identifier = "my-account-id"
#     database           = "database"
#     schema             = "schema"
#     warehouse          = "warehouse"
#     timezone_data = {
#       timezone   = "UTC"
#       utc_offset = "(UTC+00:00)"
#     }
#   }


# }

data "sifflet_datasources" "example" {}

output "datasources" {
  value = data.sifflet_datasources.example
}
