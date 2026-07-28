package user_test

import (
	"fmt"
	"regexp"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccUserResourceBasic(t *testing.T) {
	userEmail := providertests.RandomEmail()

	// All tenants have by default a domain named "All" with this static ID.
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
					resource.TestCheckTypeSetElemAttr("sifflet_user.test", "auth_types.*", "SAML2"),
					resource.TestCheckTypeSetElemNestedAttrs("sifflet_user.test", "permissions.*", map[string]string{
						"domain_id":   domainId,
						"domain_role": "VIEWER",
					}),
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
					resource.TestCheckTypeSetElemAttr("sifflet_user.test", "auth_types.*", "SAML2"),
					resource.TestCheckTypeSetElemNestedAttrs("sifflet_user.test", "permissions.*", map[string]string{
						"domain_id":   domainId,
						"domain_role": "EDITOR",
					}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_user.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
						resource "sifflet_user" "test" {
							email = "%s"
							name = "Updated name"
							role = "ADMIN"
						}
						`, userEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "name", "Updated name"),
					resource.TestCheckResourceAttr("sifflet_user.test", "role", "ADMIN"),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckNoResourceAttr("sifflet_user.test", "permissions.#"),
					resource.TestCheckTypeSetElemAttr("sifflet_user.test", "auth_types.*", "SAML2"),
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

func TestAccUserResourceAuthTypesAttributes(t *testing.T) {
	userEmail := providertests.RandomEmail()

	// All tenants have by default a domain named "All" with this static ID.
	// This is suitable for testing purposes (this domain will be present in any newly created tenant, and
	// is also present in QA tenants).
	allDomainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

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
						`, userEmail, allDomainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "email", userEmail),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.#", "2"),
					resource.TestCheckTypeSetElemAttr("sifflet_user.test", "auth_types.*", "LOGIN_PASSWORD"),
					resource.TestCheckTypeSetElemAttr("sifflet_user.test", "auth_types.*", "SAML2"),
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
						`, userEmail, allDomainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "name", "Updated name"),
					resource.TestCheckResourceAttr("sifflet_user.test", "role", "EDITOR"),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "auth_types.#", "1"),
					resource.TestCheckTypeSetElemAttr("sifflet_user.test", "auth_types.*", "LOGIN_PASSWORD"),
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

func TestAccUserResourcePermissionsAttributes(t *testing.T) {
	// Set up for creating a domain
	ctx := t.Context()
	client, err := providertests.ClientForTests(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	workspaceName := providertests.RandomName()
	assetUri := providertests.RandomGithubDeclaredAssetUri()
	assetDescription := "Created by Terraform provider tests"
	assetName := providertests.SessionPrefix() + " " + assetUri
	subTypeName := "TerraformTest"

	domainName := providertests.RandomName()

	// All tenants have by default a domain named "All" with this static ID.
	// This is suitable for testing purposes (this domain will be present in any newly created tenant, and
	// is also present in QA tenants).
	allDomainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	userEmail := providertests.RandomEmail()

	// domainConfig is reused across steps as the domain resource itself never changes.
	domainConfig := fmt.Sprintf(`
		resource "sifflet_domain" "test" {
			name = "%s"
			description = "Created by Terraform provider tests"
			dynamic_content_definition = {
				logical_operator = "AND"
				conditions = [{
					logical_operator = "IS"
					schema_uris = ["github://github.com/%s"]
				}]
			}
		}
	`, domainName, providertests.SessionPrefix())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// Create the declared assets
			asset := sifflet.PublicDeclarativeAssetDto{
				Uri:         assetUri,
				Description: &assetDescription,
				Name:        &assetName,
				Type:        sifflet.Generic,
				SubType:     &subTypeName,
			}
			err := providertests.CreateDeclaredAssets(ctx, client, workspaceName, &[]sifflet.PublicDeclarativeAssetDto{asset})
			if err != nil {
				t.Fatalf("Failed to create declared assets: %v", err)
			}
		},
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + domainConfig + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "VIEWER"
						permissions = [{
							domain_id = "%s"
							domain_role = "VIEWER"
						}, {
							domain_id = sifflet_domain.test.id
							domain_role = "EDITOR"
						}]
					}
					`, userEmail, allDomainId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "email", userEmail),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "role", "VIEWER"),
					resource.TestCheckResourceAttr("sifflet_user.test", "permissions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("sifflet_user.test", "permissions.*", map[string]string{
						"domain_id":   allDomainId,
						"domain_role": "VIEWER",
					}),
				),
			},
			{
				ResourceName:                         "sifflet_user.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				// Change order of permissions to check that it has no effect (permissions is a set)
				Config: providertests.ProviderConfig() + domainConfig + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "VIEWER"
						permissions = [{
							domain_id = sifflet_domain.test.id
							domain_role = "EDITOR"
						}, {
							domain_id = "%s"
							domain_role = "VIEWER"
						}]
					}
					`, userEmail, allDomainId),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_user.test", plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: providertests.ProviderConfig() + domainConfig + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "ADMIN"
					}
					`, userEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_user.test", "email", userEmail),
					resource.TestCheckResourceAttrSet("sifflet_user.test", "id"),
					resource.TestCheckResourceAttr("sifflet_user.test", "role", "ADMIN"),
					resource.TestCheckNoResourceAttr("sifflet_user.test", "permissions.#"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_user.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			// Delete the declared assets and all related resources
			err := providertests.DeleteDeclaredAssets(ctx, client, workspaceName)
			return err
		},
	})
}

func TestAccUserInvalidConfig(t *testing.T) {
	userEmail := providertests.RandomEmail()

	allDomainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// No permissions provided on non-ADMIN user
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "VIEWER"
					}
				`, userEmail),
				ExpectError: regexp.MustCompile("permissions must be set for non-ADMIN users"),
			},
			{
				// Empty permissions provided on non-ADMIN user
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "VIEWER"
						permissions = []
					}
				`, userEmail),
				ExpectError: regexp.MustCompile("permissions must be set for non-ADMIN users"),
			},
			{
				// Permissions provided on ADMIN user
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "ADMIN"
						permissions = [{
							domain_id = "%s"
							domain_role = "VIEWER"
						}]
					}
				`, userEmail, allDomainId),
				ExpectError: regexp.MustCompile("permissions must not be set for ADMIN users"),
			},
			{
				// Empty auth_types provided
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_user" "test" {
						email = "%s"
						name = "Terraform Test User"
						role = "ADMIN"
						auth_types = []
					}
				`, userEmail),
				ExpectError: regexp.MustCompile("Attribute auth_types set must contain at least 1 elements, got: 0"),
			},
		},
	})
}

// TestAccUserForEachPermissions tests that when "permissions" is populated dynamically via
// count/count.index the permissionsAdminValidator doesn't incorrectly report
// "permissions must be set for non-ADMIN users".
//
// This uses `count` rather than `for_each` because the acceptance testing framework doesn't
// support for_each-indexed resources in test configs: https://github.com/hashicorp/terraform-plugin-sdk/issues/536
// `count` still exercises the same code path, since the computed permissions value is unknown
// at plan time either way.
func TestAccUserForEachPermissions(t *testing.T) {
	userEmail := providertests.RandomEmail()

	// All tenants have by default a domain named "All" with this static ID.
	domainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					locals {
						viewer_users = [
							{
								email = "%s"
								permissions = [{
									domain_id   = "%s"
									domain_role = "VIEWER"
								}]
							}
						]
					}

					resource "sifflet_user" "test" {
						count       = length(local.viewer_users)
						email       = local.viewer_users[count.index].email
						name        = "Terraform Test User"
						role        = "VIEWER"
						permissions = local.viewer_users[count.index].permissions
					}
					`, userEmail, domainId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sifflet_user.test[0]",
						tfjsonpath.New("email"),
						knownvalue.StringExact(userEmail),
					),
					statecheck.ExpectKnownValue(
						"sifflet_user.test[0]",
						tfjsonpath.New("permissions").AtSliceIndex(0).AtMapKey("domain_id"),
						knownvalue.StringExact(domainId),
					),
				},
			},
		},
	})
}
