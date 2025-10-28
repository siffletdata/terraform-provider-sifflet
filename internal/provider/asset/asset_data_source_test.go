package asset_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAssetDataSource(t *testing.T) {
	// We need to use the workspace sync endpoint to create assets, so that we can test the asset data source.
	// Usually, assets would be created through a source ingestion, but we do not want to use real sources for tests.
	ctx := context.Background()
	dryRun := false
	client, err := providertests.ClientForTests(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	workspaceName := providertests.RandomName()
	assetUri := providertests.RandomGithubDeclaredAssetUri()
	assetDescription := "Created by Terraform provider tests"
	assetName := providertests.SessionPrefix() + " " + assetUri
	subTypeName := "TerraformTest"

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
			payload := sifflet.PublicDeclarativePayloadDto{
				Assets:    &[]sifflet.PublicDeclarativeAssetDto{asset},
				Workspace: workspaceName,
			}
			params := sifflet.PublicSyncAssetsParams{
				DryRun: &dryRun,
			}

			response, err := client.PublicSyncAssetsWithResponse(ctx, &params, payload)
			if err != nil {
				t.Fatalf("Failed to sync assets: %v", err)
			}
			if response.StatusCode() != http.StatusOK {
				t.Fatalf("Failed to sync assets: status code %d. Details: %s", response.StatusCode(), response.Body)
			}
		},
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_asset" "test" {
					uri = "%s"
				}`, assetUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_asset.test", "uri", assetUri),
					resource.TestCheckResourceAttr("data.sifflet_asset.test", "name", assetName),
					resource.TestCheckResourceAttr("data.sifflet_asset.test", "type", "OTHER"),
					resource.TestCheckResourceAttr("data.sifflet_asset.test", "uri", assetUri),
					resource.TestCheckResourceAttrSet("data.sifflet_asset.test", "id"),
					resource.TestCheckResourceAttr("data.sifflet_asset.test", "description", "Created by Terraform provider tests"),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			// Delete the declared assets and all related resources
			response, err := client.PublicDeleteWorkspaceWithResponse(ctx, workspaceName, &sifflet.PublicDeleteWorkspaceParams{DryRun: &dryRun})
			if err != nil {
				return fmt.Errorf("Failed to delete workspace: %v", err)
			}
			if response.StatusCode() != http.StatusOK {
				return fmt.Errorf("Failed to delete workspace: status code %d. Details: %s", response.StatusCode(), response.Body)
			}
			return nil
		},
	})
}

func TestAccAssetReadErrorDataSource(t *testing.T) {
	assetUri := "snowflake://sifflet-internal/DEMO.TEST_ONLY.DOES_NOT_EXIST"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_asset" "test" {
					uri = "%s"
				}`, assetUri),
				ExpectError: regexp.MustCompile("HTTP status code: 403. Details: Access Denied"),
			},
		},
	})
}
