package provider

import (
	"fmt"
	"regexp"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSourcesDataSource(t *testing.T) {

	sourceName := providertests.RandomName()
	tagName := providertests.RandomName()
	databaseName := providertests.RandomName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// No result
			{
				Config: providertests.ProviderConfig() + `
				data "sifflet_sources" "test" {
					filter = {
						text_search = "doesn't match any data source"
						types = ["MYSQL"]
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_sources.test", "results.#", "0"),
				),
			},
			// With result
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				resource "sifflet_tag" "test" {
					name = "%s"
					description = "Terraform acceptance tests."
				}

				resource "sifflet_source" "test" {
					name = "%s"
					parameters = {
						athena = {
							database = "%s"
							datasource = "datasource"
							region = "region"
							role_arn = "arn:aws:iam::123456789012:role/role"
							s3_output_location = "s3://mybucket"
							workgroup = "workgroup"
						}
					}
					tags = [{name = sifflet_tag.test.name}]
				}

				data "sifflet_sources" "test" {
					filter = {
						text_search = sifflet_source.test.name
						types = ["ATHENA"]
						tags = [{ name = sifflet_tag.test.name}]
					}
				}`, tagName, sourceName, databaseName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_sources.test", "results.#", "1"),
					resource.TestCheckResourceAttr("data.sifflet_sources.test", "results.0.name", sourceName),
				),
			},
		},
	})
}

func TestAccSourcesErrorDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + `
				data "sifflet_sources" "test" {
					filter = {
						text_search = "test"
						types = ["MYSQL"]
						tags = [ { name = "a tag that doesn't exist", kind = "Classification" } ]
					}
				}`,
				ExpectError: regexp.MustCompile("HTTP status code: 400. Details: Tag not found"),
			},
			{
				Config: providertests.ProviderConfig() + `
				data "sifflet_sources" "test" {
					filter = {
						types = ["not_a_valid_type"]
					}
				}`,
				ExpectError: regexp.MustCompile("HTTP status code: 400. Details: filter.types"),
			},
		},
	})
}
