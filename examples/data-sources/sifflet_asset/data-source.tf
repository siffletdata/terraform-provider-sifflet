data "sifflet_asset" "example" {
  uri = "scheme://authority/unique.name"
}

output "asset_urn" {
  value = data.sifflet_asset.example.urn
}
