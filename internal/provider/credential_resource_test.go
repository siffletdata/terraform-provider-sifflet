package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialResourceBasic(t *testing.T) {
	credentialName := RandomName()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
						resource "sifflet_credential" "test" {
							name = "testterraform"
							description = "A description"
							value = "%s"
						}
						`, credentialName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credential.test", "name", "testterraform"),
					resource.TestCheckResourceAttr("sifflet_credential.test", "description", "A description"),
					resource.TestCheckResourceAttr("sifflet_credential.test", "value", "A value"),
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
		},
	})
}
