data "sifflet_assets" "test" {
  filter = {
    text_search     = "asset_name"
    type_categories = ["TABLE_AND_VIEW"]
    tags = [{
      name = "tag_name"
    }]
    max_results = 10
  }
}

output "assets" {
  value = data.sifflet_assets.test.results
}
