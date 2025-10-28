package source_v2_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSourceV2SchemasDataSource(t *testing.T) {
	// We need to use the source v1 sync endpoint to create a sub-source, so that we can test the source_v2_schemas data source.
	// Usually, sub-sources would be created through a source ingestion, but we do not want to use real sources for tests,
	// and we cannot reliably wait for a source ingestion to complete before trying to get the sub-sources.
	ctx := t.Context()
	client, err := providertests.ClientForTests(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create the source v1 (sub-source)
	subSourceName := providertests.RandomName()
	subSourceDescription := "Created by Terraform provider tests"
	accountId := providertests.RandomName()
	awsRegion := "eu-west-1"
	roleArn := "arn:aws:iam::123456789012:role/role"
	// We use a quicksight source because it does not require secrets, so it is easier to setup.
	dto := sifflet.PublicQuicksightParametersDto{
		Type:      sifflet.PublicQuicksightParametersDtoTypeQUICKSIGHT,
		AccountId: &accountId,
		AwsRegion: &awsRegion,
		RoleArn:   &roleArn,
	}
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	err = parametersDto.FromPublicQuicksightParametersDto(dto)
	if err != nil {
		t.Fatalf("Failed to create parameters: %v", err)
	}
	source := sifflet.PublicCreateSourceDto{
		Name:        subSourceName,
		Description: &subSourceDescription,
		Parameters:  parametersDto,
	}
	sourceResponse, err := client.PublicCreateSourceWithResponse(ctx, source)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}
	if sourceResponse.StatusCode() != http.StatusCreated {
		t.Fatalf("Failed to create source: status code %d", sourceResponse.StatusCode())
	}

	// Get the source v2 id corresponding to the created sub-source
	// The source v2 getAll endpoint does not allow filtering by name, so we need to iterate over all sources.
	var sourceId string
	sourcesResponse, err := client.PublicGetSourcesV2WithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to get sources: %v", err)
	}
	if sourcesResponse.StatusCode() != http.StatusOK {
		t.Fatalf("Failed to get sources: status code %d", sourcesResponse.StatusCode())
	}
	sources := sourcesResponse.JSON200.Data
	if len(sources) == 0 {
		t.Fatalf("No sources found")
	}
	for _, source := range sources {
		var sourceItem sifflet.PublicGetSourceV2DataItem
		err := sourceItem.FromPublicPageDtoPublicGetSourceV2DtoDataItem(source)
		if err != nil {
			t.Fatalf("Failed to get source: %v", err)
		}
		if strings.Contains(*sourceItem.Name, providertests.SessionPrefix()) {
			sourceId = *sourceItem.Id
			break
		}
	}
	if sourceId == "" {
		t.Fatalf("No source found with name containing %s", providertests.SessionPrefix())
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_source_v2_schemas" "test" {
					source_id = "%s"
				}`, sourceId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_source_v2_schemas.test", "schemas.#", "1"),
					resource.TestCheckResourceAttr("data.sifflet_source_v2_schemas.test", "schemas.0.uri", fmt.Sprintf("quicksight://%s.eu-west-1.amazonaws.com", accountId)),
				),
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			// Delete the source v2, which will also delete the source v1 (sub-source)
			uuidSourceId, err := uuid.Parse(sourceId)
			if err != nil {
				t.Fatalf("Failed to parse sourceId as UUID: %v", err)
			}
			sourceResponse, err := client.PublicDeleteSourceV2WithResponse(ctx, uuidSourceId)
			if err != nil {
				t.Fatalf("Failed to delete source: %v", err)
			}
			if sourceResponse.StatusCode() != http.StatusNoContent {
				t.Fatalf("Failed to delete source: status code %d", sourceResponse.StatusCode())
			}
			return nil
		},
	})
}

func TestAccSourceV2SchemasReadErrorDataSource(t *testing.T) {
	assetId := "00000000-0000-0000-0000-000000000000"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_source_v2_schemas" "test" {
					source_id = "%s"
				}`, assetId),
				ExpectError: regexp.MustCompile("HTTP status code: 404. Details: Source not found"),
			},
		},
	})
}
