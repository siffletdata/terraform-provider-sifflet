resource "sifflet_tag" "super_tag" {
  name        = "my-tag"
  description = "My super tag."
}

resource "sifflet_datasource" "bigquery" {
  name            = "orders"
  secret_id       = "bq_secrets"
  cron_expression = "@daily"
  bigquery = {
    billing_project_id = "billing-prj"
    dataset_id         = "orders"
    project_id         = "orders-prj"
  }

  tags = [
    sifflet_tag.super_tag.id
  ]
}


resource "sifflet_datasource" "dbt" {
  name            = "orders-model"
  cron_expression = "@daily"
  dbt = {
    project_name = "orders"
    target       = "prod"
  }
}

resource "sifflet_datasource" "snowflake" {
  name            = "snwoflake-orders"
  secret_id       = "snowflake_creds"
  cron_expression = "@daily"
  snowflake = {
    account_identifier = "my-account-id"
    database           = "database"
    schema             = "schema"
    warehouse          = "warehouse"
    timezone_data = {
      timezone   = "UTC"
      utc_offset = "(UTC+00:00)"
    }
  }
}

resource "sifflet_datasource" "power_bi" {
  name            = "powerbi-orders"
  secret_id       = "powerbi_creds"
  cron_expression = "@daily"

  power_bi = {
    tenant_id    = "tenant-id"
    client_id    = "client-id"
    workspace_id = "workspace-id"

    timezone_data = {
      timezone   = "UTC"
      utc_offset = "(UTC+00:00)"
    }
  }
}
