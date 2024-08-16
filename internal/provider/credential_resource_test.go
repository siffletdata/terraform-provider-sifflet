package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialResourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// FIXME: random name to avoid conflicts
				Config: providerConfig + `
						resource "sifflet_credential" "test" {
							name = "testterraform"
							description = "A description"
							value = "A value"
						}
						`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_credential.test", "name", "testterraform"),
					resource.TestCheckResourceAttr("sifflet_credential.test", "description", "A description"),
					resource.TestCheckResourceAttr("sifflet_credential.test", "value", "A value"),
				),
			},
		},
	})
}
