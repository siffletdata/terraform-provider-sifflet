package credentials_test

import (
	"fmt"
	"regexp"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialDataSource(t *testing.T) {
	credentialsName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				resource "sifflet_credentials" "test" {
					name = "%s"
					description = "Description"
					value = "Value"
				}

				data "sifflet_credentials" "test" {
					name = sifflet_credentials.test.name
				}`, credentialsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_credentials.test", "name", credentialsName),
					resource.TestCheckResourceAttr("data.sifflet_credentials.test", "description", "Description"),
				),
			},
		},
	})
}

func TestAccCredentialReadErrorDataSource(t *testing.T) {
	credentialsName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_credentials" "test" {
					name = "%s"
				}`, credentialsName),
				ExpectError: regexp.MustCompile("HTTP status code: 404. Details: Credentials not found:"),
			},
		},
	})
}
