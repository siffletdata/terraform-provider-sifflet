package domain_test

import (
	"fmt"
	"regexp"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSourceInvalidConfig(t *testing.T) {
	domainName := providertests.RandomName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
					}
				`, domainName),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = []
						}
						static_content_definition = {
							asset_uris = []
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						static_content_definition = {
							asset_uris = []
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("must contain at least 1"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = []
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("must contain at least 1"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = [{
								logical_operator = "IS"
							}]
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("No attribute specified when one"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = [{
								logical_operator = "IS"
								schema_uris = []
								tags = []
							}]
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("2 attributes specified when one"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = [{
								logical_operator = "IS"
								tags = [{
									kind = "SNOWFLAKE_EXTERNAL"
								}]
							}]
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("No attribute specified when one"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = [{
								logical_operator = "IS"
								tags = [{
									name = "SNOWFLAKE_EXTERNAL"
									id = "00000000-0000-0000-0000-000000000000"
								}]
							}]
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("2 attributes specified when one"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "AND"
							conditions = [{
								logical_operator = "IS"
								tags = [{
									kind = "SNOWFLAKE_EXTERNAL"
									id = "00000000-0000-0000-0000-000000000000"
								}]
							}]
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
					resource "sifflet_domain" "test" {
						name = "%s"
						dynamic_content_definition = {
							logical_operator = "XOR"
							conditions = [{
								logical_operator = "IS"
								tags = [{
									kind = "SNOWFLAKE_EXTERNAL"
									name = "test"
								}]
							}]
						}
					}
				`, domainName),
				ExpectError: regexp.MustCompile("logical_operator value must be one of"),
			},
		},
	})
}
func TestAccDomainStaticContentDefinitionResource(t *testing.T) {
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
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_asset" "test" {
					uri = "%s"
				}

				resource "sifflet_domain" "test" {
					name = "%s"
					description = "Created by Terraform provider tests"
					static_content_definition = {
						asset_uris = [data.sifflet_asset.test.uri]
					}
				}
				`, assetUri, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("sifflet_domain.test", "description", "Created by Terraform provider tests"),
					resource.TestCheckResourceAttr("sifflet_domain.test", "static_content_definition.asset_uris.0", assetUri),
				),
			},
			{
				ResourceName:                         "sifflet_domain.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_asset" "test" {
					uri = "%s"
				}

				resource "sifflet_domain" "test" {
					name = "%s-updated"
					description = "Created by Terraform provider tests - updated"
					static_content_definition = {
						asset_uris = [data.sifflet_asset.test.uri]
					}
				}
				`, assetUri, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_domain.test", "name", fmt.Sprintf("%s-updated", domainName)),
					resource.TestCheckResourceAttr("sifflet_domain.test", "description", "Created by Terraform provider tests - updated"),
					resource.TestCheckResourceAttr("sifflet_domain.test", "static_content_definition.asset_uris.0", assetUri),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			// Delete the declared assets and all related resources
			err := providertests.DeleteDeclaredAssets(ctx, client, workspaceName)
			return err
		},
	})
}

func TestAccDomainDynamicContentDefinitionResource(t *testing.T) {
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
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
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
				`, domainName, providertests.SessionPrefix()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("sifflet_domain.test", "description", "Created by Terraform provider tests"),
					resource.TestCheckResourceAttr("sifflet_domain.test", "dynamic_content_definition.logical_operator", "AND"),
					resource.TestCheckResourceAttr("sifflet_domain.test", "dynamic_content_definition.conditions.0.logical_operator", "IS"),
					resource.TestCheckResourceAttr("sifflet_domain.test", "dynamic_content_definition.conditions.0.schema_uris.0", fmt.Sprintf("github://github.com/%s", providertests.SessionPrefix())),
					resource.TestCheckNoResourceAttr("sifflet_domain.test", "dynamic_content_definition.conditions.0.tags"),
					resource.TestCheckNoResourceAttr("sifflet_domain.test", "static_content_definition"),
				),
			},
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				resource "sifflet_domain" "test" {
					name = "%s"
					description = "Created by Terraform provider tests"
					dynamic_content_definition = {
						logical_operator = "AND"
						conditions = [{
							logical_operator = "IS"
							schema_uris = ["github://github.com/%s"]
						},
						{
							logical_operator = "IS"
							tags = [{
								kind = "SNOWFLAKE_EXTERNAL"
								name = "test"
							}]
						}]
					}
				}
				`, domainName, providertests.SessionPrefix()),
				ExpectError: regexp.MustCompile("HTTP status code: 400. Details: Tag not found"),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			// Delete the declared assets and all related resources
			err := providertests.DeleteDeclaredAssets(ctx, client, workspaceName)
			return err
		},
	})
}
