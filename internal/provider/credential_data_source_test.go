package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialDataSource(t *testing.T) {
	credentialName := randomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "sifflet_credential" "test" {
					name = "%s"
					description = "Description"
					value = "Value"
				}

				data "sifflet_credential" "test" {
					name = sifflet_credential.test.name
				}`, credentialName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_credential.test", "name", credentialName),
					resource.TestCheckResourceAttr("data.sifflet_credential.test", "description", "Description"),
				),
			},
		},
	})
}

func TestAccCredentialReadErrorDataSource(t *testing.T) {
	credentialName := randomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
				data "sifflet_credential" "test" {
					name = "%s"
				}`, credentialName),
				ExpectError: regexp.MustCompile("HTTP status code: 404. Details: Credentials not found:"),
			},
		},
	})
}
