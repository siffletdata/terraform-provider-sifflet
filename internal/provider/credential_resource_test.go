package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialResourceBasic(t *testing.T) {
	credentialName := strings.ReplaceAll(RandomName(), "-", "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
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
				Config: providerConfig + fmt.Sprintf(`
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
