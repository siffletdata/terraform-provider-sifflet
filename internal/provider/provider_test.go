package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestProviderInvalidConfigs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "sifflet" {
					// This is not a valid Sifflet instance URL
					host = "https://siffletdata-status.com"
					token = "my-token"
				}

				// Dummy resource to force the provider instantiation.
				data "sifflet_credentials" "test" {
					name = "my-credential"
				}
				`,
				ExpectError: regexp.MustCompile("Got an unexpected status code when attempting to perform a health check"),
			},
			{
				Config: `
				provider "sifflet" {
					// Use the SIFFLET_HOST environment variable (assumed to be always set during tests),
					// but set an invalid token
					token = "invalid token"
				}

				data "sifflet_credentials" "test" {
					name = "my-credential"
				}
				`,
				ExpectError: regexp.MustCompile("HTTP status code: 401. Details: Full authentication is required to access"),
			},
		},
	})
}
