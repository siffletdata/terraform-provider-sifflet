# The credentials value must match what the source expects.
# See the Sifflet documentation for each source type for details
# on the credential value format.
data "sifflet_credentials" "example" {
  name = "example"
}

# A simple BigQuery data source.
resource "sifflet_source" "example" {
  name        = "example"
  description = "A description"
  credentials = sifflet_credentials.example.name
  parameters = {
    # Pass the parameter block that matches the source type.
    bigquery = {
      project_id         = "project_id"
      dataset_id         = "dataset"
      billing_project_id = "dataset"
    }
  }
}

# Example with more complex parameters.
resource "sifflet_source" "complex" {
  name        = "example"
  description = "A description"
  credentials = sifflet_credentials.example.name
  parameters = {
    snowflake = {
      account_identifier = "accountidentifier"
      database           = "database"
      schema             = "schema"
      warehouse          = "warehouse"
    }
  }
  schedule = "@daily"
  # The timezone can also be a timezone name, e.g. "Europe/Paris".
  timezone = "UTC+1"
  tags = [{
    # Tags specified this way must be created before the source.
    name = "tag_name"
  }]
}
