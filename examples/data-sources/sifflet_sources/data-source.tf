data "sifflet_sources" "test" {
  filter = {
    text_search = "source_name"
    types       = ["MYSQL"]
    tags = [{
      name = "tag_name"
    }]
  }
}
