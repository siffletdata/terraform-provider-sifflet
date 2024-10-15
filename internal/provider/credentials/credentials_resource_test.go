package credentials_test

import (
	"fmt"
	"regexp"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialResourceBasic(t *testing.T) {
	credentialsName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credentials" "test" {
							name = "%s"
							description = "A description"
							value = "Secret value"
						}
						`, credentialsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credentials.test", "name", credentialsName),
					resource.TestCheckResourceAttr("sifflet_credentials.test", "description", "A description"),
					resource.TestCheckResourceAttr("sifflet_credentials.test", "value", "Secret value"),
				),
			},
			{
				ResourceName:                         "sifflet_credentials.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        credentialsName,
				ImportStateVerifyIdentifierAttribute: "name",
				// The API never returns the secret value so we can't import it
				ImportStateVerifyIgnore: []string{"value"},
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credentials" "test" {
							name = "%s"
							description = "An updated description"
							value = "Updated value"
						}
						`, credentialsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credentials.test", "name", credentialsName),
					resource.TestCheckResourceAttr("sifflet_credentials.test", "description", "An updated description"),
					resource.TestCheckResourceAttr("sifflet_credentials.test", "value", "Updated value"),
				),
			},
		},
	})
}

func TestAccCredentialNoValue(t *testing.T) {
	credentialsName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credentials" "test" {
							name = "%s"
							description = "A description"
							// Value purposely omitted, this is an error when creating or importing credentials
						}
						`, credentialsName),
				ExpectError: regexp.MustCompile("The value attribute is required when creating credentials"),
			},
			// Create the credentials resource
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credentials" "test" {
							name = "%s"
							description = "A description"
							value = "Secret"
						}
						`, credentialsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credentials.test", "value", "Secret"),
				),
			},
			// Now, ensure its description can be updated even if value is removed from the configuration
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credentials" "test" {
							name = "%s"
							description = "Updated description"
						}
						`, credentialsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credentials.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccCredentialInvalidName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + `
						resource "sifflet_credentials" "test" {
							name = "invalid-name-123"
							description = "A description"
							value = "Secret value"
						}
						`,
				ExpectError: regexp.MustCompile("Attribute name must start and end with a letter"),
			},
		},
	})
}
