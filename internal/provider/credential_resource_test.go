package provider

import (
	"fmt"
	"regexp"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialResourceBasic(t *testing.T) {
	credentialName := providertests.RandomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credential" "test" {
							name = "%s"
							description = "A description"
							value = "Secret value"
						}
						`, credentialName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credential.test", "name", credentialName),
					resource.TestCheckResourceAttr("sifflet_credential.test", "description", "A description"),
					resource.TestCheckResourceAttr("sifflet_credential.test", "value", "Secret value"),
				),
			},
			{
				ResourceName:                         "sifflet_credential.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        credentialName,
				ImportStateVerifyIdentifierAttribute: "name",
				// The API never returns the secret value so we can't import it
				ImportStateVerifyIgnore: []string{"value"},
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credential" "test" {
							name = "%s"
							description = "An updated description"
							value = "Updated value"
						}
						`, credentialName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credential.test", "name", credentialName),
					resource.TestCheckResourceAttr("sifflet_credential.test", "description", "An updated description"),
					resource.TestCheckResourceAttr("sifflet_credential.test", "value", "Updated value"),
				),
			},
		},
	})
}

func TestAccCredentialNoValue(t *testing.T) {
	credentialName := providertests.RandomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credential" "test" {
							name = "%s"
							description = "A description"
							// Value purposely omitted, this is an error when creating or importing a credential
						}
						`, credentialName),
				ExpectError: regexp.MustCompile("The value attribute is required when creating a credential"),
			},
			// Create the credential resource
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credential" "test" {
							name = "%s"
							description = "A description"
							value = "Secret"
						}
						`, credentialName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credential.test", "value", "Secret"),
				),
			},
			// Now, ensure its description can be updated even if value is removed from the configuration
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_credential" "test" {
							name = "%s"
							description = "Updated description"
						}
						`, credentialName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credential.test", "description", "Updated description"),
				),
			},
		},
	})
}
