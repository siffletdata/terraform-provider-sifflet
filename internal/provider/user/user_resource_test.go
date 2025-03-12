package user_test

import (
	"fmt"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccUserResourceBasic(t *testing.T) {
	userEmail := providertests.RandomEmail()
	// As of this writing, there's no public API for domains.
	// However, all tenants have by default a domain named "All" with this static ID.
	// This is suitable for testing purposes (this domain will be present in any newly created tenant, and
	// is also present in QA tenants).
	domainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test" {
							email = "%s"
							name = "Terraform Test User"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}
						`, userEmail, domainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "email", userEmail),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.0", "SAML2"),
				),
			},
			{
				ResourceName:                         "sifflet_user.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test" {
							email = "%s"
							name = "Updated name"
							role = "EDITOR"
							permissions = [{
								domain_id = "%s"
								domain_role = "EDITOR"
							}]
						}
						`, userEmail, domainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "name", "Updated name"),
					resource.TestCheckResourceAttr("sifflet_user.test", "role", "EDITOR"),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.0", "SAML2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_user.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccUserResourceOptionalAttributes(t *testing.T) {
	userEmail := providertests.RandomEmail()
	// As of this writing, there's no public API for domains.
	// However, all tenants have by default a domain named "All" with this static ID.
	// This is suitable for testing purposes (this domain will be present in any newly created tenant, and
	// is also present in QA tenants).
	domainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test" {
							email = "%s"
							name = "Terraform Test User"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
							// These are set in inverse alphabetical order because the backend sorts them and we need to check
							// that this does not introduce a bug
							auth_types = ["SAML2", "LOGIN_PASSWORD"]
						}
						`, userEmail, domainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "email", userEmail),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.0", "LOGIN_PASSWORD"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.1", "SAML2"),
				),
			},
			{
				ResourceName:                         "sifflet_user.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test" {
							email = "%s"
							name = "Updated name"
							role = "EDITOR"
							permissions = [{
								domain_id = "%s"
								domain_role = "EDITOR"
							}]
							auth_types = ["LOGIN_PASSWORD"]
						}
						`, userEmail, domainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "name", "Updated name"),
					resource.TestCheckResourceAttr("sifflet_user.test", "role", "EDITOR"),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.0", "LOGIN_PASSWORD"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_user.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
