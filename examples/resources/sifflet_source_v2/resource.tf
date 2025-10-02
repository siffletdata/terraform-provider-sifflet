# The format of the credentials value must match what the source expects.
# See the Sifflet documentation for each source type for details
# about the credential value format.
data "sifflet_credentials" "example" {
  name = "example"
}

# A simple BigQuery data source.
resource "sifflet_source_v2" "example" {
  name = "example"
  parameters = {
    # Pass the parameter block that matches the source type.
    bigquery = {
      project_id         = "project_id"
      billing_project_id = "billing_project_id"
      credentials        = sifflet_credentials.example.name
    }
  }
}

# Example with a schedule.
resource "sifflet_source_v2" "scheduled" {
  name = "example"
  parameters = {
    snowflake = {
      account_identifier = "accountidentifier"
      warehouse          = "warehouse"
      credentials        = sifflet_credentials.example.name
      schedule           = "@daily"
    }
  }
}
