package provider

import (
	"fmt"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagResourceBasic(t *testing.T) {
	tagName := providertests.RandomName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_tag" "test" {
							name = "%s"
							description = "A description"
						}
						`, tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("sifflet_tag.test", "description", "A description"),
				),
			},
			{
				ResourceName:      "sifflet_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_tag" "test" {
							name = "%s"
							description = "An updated description"
						}
						`, tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_tag.test", "name", tagName),
					resource.TestCheckResourceAttr("sifflet_tag.test", "description", "An updated description"),
				),
			},
		},
	})
}
