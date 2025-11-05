# --- Example of a domain with static content definition ---

# The asset URIs can be found in the Sifflet UI
data "sifflet_asset" "example_by_uri" {
  uri = "scheme://authority/unique.name"
}

# You can also use the `sifflet_assets` data source to search assets using filters
data "sifflet_assets" "example" {
  filter = {
    text_search     = "asset_name"
    type_categories = ["TABLE_AND_VIEW"]
    tags = [{
      name = "tag_name"
    }]
  }
}

resource "sifflet_domain" "static" {
  name        = "Static example"
  description = "Static domain example"
  static_content_definition = {
    asset_uris = concat(
      [for asset in data.sifflet_assets.example.results : asset.uri],
      [data.sifflet_asset.example_by_uri.uri]
    )
  }
}

# --- Example of a domain with dynamic content definition ---

# The schema URIs can be constructed using the documentation: https://docs.siffletdata.com/docs/uri
# You can also use the `sifflet_source_v2_schemas` data source to get all schemas contained in a source
data "sifflet_source_v2_schemas" "example" {
  source_id = "00000000-0000-0000-0000-000000000000"
}

resource "sifflet_domain" "example" {
  name        = "Dynamic example"
  description = "Dynamic domain example"
  dynamic_content_definition = {
    logical_operator = "AND"
    conditions = [{
      logical_operator = "IS_NOT"
      tags = [{
        kind = "SNOWFLAKE_EXTERNAL"
        name = "prod"
      }]
      }, {
      logical_operator = "IS"
      schema_uris      = [for schema in data.sifflet_source_v2_schemas.example.schemas : schema.uri]
    }]
  }
}
