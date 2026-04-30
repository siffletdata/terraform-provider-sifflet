package team_test

import (
	"fmt"
	"testing"

	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTeamResourceBasic(t *testing.T) {
	teamName := providertests.RandomName()

	// All tenants have by default a domain named "All" with this static ID.
	// This is suitable for testing purposes (this domain will be present in any newly created tenant, and
	// is also present in QA tenants).
	domainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_team" "test" {
							name = "%s"
							description = "Test team created by Terraform"
							domain_permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}
						`, teamName, domainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_team.test", "name", teamName),
					resource.TestCheckResourceAttr("sifflet_team.test", "description", "Test team created by Terraform"),
					resource.TestCheckResourceAttrSet("sifflet_team.test", "id"),
					resource.TestCheckTypeSetElemNestedAttrs("sifflet_team.test", "domain_permissions.*", map[string]string{
						"domain_id":   domainId,
						"domain_role": "VIEWER",
					}),
				),
			},
			{
				ResourceName:                         "sifflet_team.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_team" "test" {
							name = "%s Updated"
							description = "Updated description"
							domain_permissions = [{
								domain_id = "%s"
								domain_role = "CATALOG_EDITOR"
							}]
						}
						`, teamName, domainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_team.test", "name", teamName+" Updated"),
					resource.TestCheckResourceAttr("sifflet_team.test", "description", "Updated description"),
					resource.TestCheckResourceAttrSet("sifflet_team.test", "id"),
					resource.TestCheckTypeSetElemNestedAttrs("sifflet_team.test", "domain_permissions.*", map[string]string{
						"domain_id":   domainId,
						"domain_role": "CATALOG_EDITOR",
					}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_team.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccTeamResourceWithUsers(t *testing.T) {
	teamName := providertests.RandomName()
	userEmail1 := providertests.RandomEmail()
	userEmail2 := providertests.RandomEmail()
	userEmail3 := providertests.RandomEmail()
	domainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test1" {
							email = "%s"
							name = "Test User 1 for Team"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}

						resource "sifflet_user" "test2" {
							email = "%s"
							name = "Test User 2 for Team"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}

						resource "sifflet_team" "test" {
							name = "%s"
							description = "Team with users"
							users = [
								{
									user_id = sifflet_user.test1.id
								},
								{
									email = sifflet_user.test2.email
								}
							]
						}
						`, userEmail1, domainId, userEmail2, domainId, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_team.test", "name", teamName),
					resource.TestCheckResourceAttr("sifflet_team.test", "description", "Team with users"),
					resource.TestCheckResourceAttrSet("sifflet_team.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("sifflet_team.test", "users.*.user_id", "sifflet_user.test1", "id"),
					resource.TestCheckTypeSetElemAttrPair("sifflet_team.test", "users.*.email", "sifflet_user.test2", "email"),
				),
			},
			{
				ResourceName:                         "sifflet_team.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test1" {
							email = "%s"
							name = "Test User 1 for Team"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}

						resource "sifflet_user" "test2" {
							email = "%s"
							name = "Test User 2 for Team"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}

						resource "sifflet_user" "test3" {
							email = "%s"
							name = "Test User 3 for Team"
							role = "VIEWER"
							permissions = [{
								domain_id = "%s"
								domain_role = "VIEWER"
							}]
						}

						resource "sifflet_team" "test" {
							name = "%s"
							description = "Team with updated users"
							users = [
								{
									user_id = sifflet_user.test2.id
								},
								{
									email = sifflet_user.test3.email
								}
							]
						}
						`, userEmail1, domainId, userEmail2, domainId, userEmail3, domainId, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_team.test", "name", teamName),
					resource.TestCheckResourceAttr("sifflet_team.test", "description", "Team with updated users"),
					resource.TestCheckResourceAttrSet("sifflet_team.test", "id"),
					resource.TestCheckTypeSetElemAttrPair("sifflet_team.test", "users.*.user_id", "sifflet_user.test2", "id"),
					resource.TestCheckTypeSetElemAttrPair("sifflet_team.test", "users.*.email", "sifflet_user.test3", "email"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_team.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
