package tag_test

import (
	"fmt"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTagResourceBasic(t *testing.T) {
	tagName := providertests.RandomName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_tag.test", plancheck.ResourceActionCreate),
					},
				},
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_tag.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
