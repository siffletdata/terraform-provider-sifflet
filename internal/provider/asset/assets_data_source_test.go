package asset_test

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

func TestAccAssetsDataSource(t *testing.T) {
	// We need to use the workspace sync endpoint to create assets, so that we can test the assets data source.
	// Usually, assets would be created through a source ingestion, but we do not want to use real sources for tests.
	ctx := t.Context()
	client, err := providertests.ClientForTests(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	workspaceName := providertests.RandomName()
	subTypeName := "TerraformTest"
	//  This tag is present on all Sifflet instances by default
	tagName := "Non-Production"
	assetUri := providertests.RandomGithubDeclaredAssetUri()
	assetDescription := "Created by Terraform provider tests"
	assetName := providertests.SessionPrefix() + " " + assetUri
	secondAssetUri := providertests.RandomGithubDeclaredAssetUri()
	secondAssetName := providertests.SessionPrefix() + " " + secondAssetUri

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// Create the declared assets
			asset := sifflet.PublicDeclarativeAssetDto{
				Uri: assetUri,
				Tags: &[]sifflet.PublicTagReferenceDto{
					{
						Name: &tagName,
					},
				},
				Description: &assetDescription,
				Name:        &assetName,
				Type:        sifflet.Generic,
				SubType:     &subTypeName,
			}
			secondAsset := sifflet.PublicDeclarativeAssetDto{
				Uri:         secondAssetUri,
				Description: &assetDescription,
				Name:        &secondAssetName,
				Type:        sifflet.Generic,
				SubType:     &subTypeName,
			}
			err := providertests.CreateDeclaredAssets(ctx, client, workspaceName, &[]sifflet.PublicDeclarativeAssetDto{asset, secondAsset})
			if err != nil {
				t.Fatalf("Failed to create declared assets: %v", err)
			}
		},
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// No result
			{
				Config: providertests.ProviderConfig() + `
				data "sifflet_assets" "test" {
					filter = {
						text_search = "doesn't match any asset"
						types = ["declared-asset_Repository"]
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.#", "0"),
				),
			},
			// With result
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_assets" "test" {
					filter = {
						text_search = "%s"
						type_categories = ["declared-asset_%s"]
						tags = [{ name = "%s" }]
					}
				}`, providertests.SessionPrefix(), subTypeName, tagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.#", "1"),
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.0.name", assetName),
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.0.type", "OTHER"),
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.0.uri", assetUri),
					resource.TestCheckResourceAttrSet("data.sifflet_assets.test", "results.0.id"),
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.0.description", assetDescription),
				),
			},
			// With multiple results
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_assets" "test" {
					filter = {
						text_search = "%s"
						type_categories = ["declared-asset_%s"]
					}
				}`, providertests.SessionPrefix(), subTypeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.#", "2"),
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.0.name", assetName),
					resource.TestCheckResourceAttr("data.sifflet_assets.test", "results.1.name", secondAssetName),
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

func TestAccAssetsReadErrorDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + `
				data "sifflet_assets" "test" {
					filter = {
						text_search = "test"
						type_categories = ["TABLE_AND_VIEW"]
						tags = [ { name = "a tag that doesn't exist", kind = "Classification" } ]
					}
				}`,
				ExpectError: regexp.MustCompile("HTTP status code: 400. Details: Tag not found"),
			},
		},
	})
}
