package provider

import (
	"fmt"
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
