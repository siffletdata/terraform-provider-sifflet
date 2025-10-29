package source_v2_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/runtime/types"
)

func TestAccSourceV2SchemasDataSource(t *testing.T) {
	ctx := t.Context()
	client, err := providertests.ClientForTests(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	accountId := providertests.RandomName()

	// We need to use the source v1 sync endpoint to create a sub-source, so that we can test the source_v2_schemas data source.
	// Usually, sub-sources would be created through a source ingestion, but we do not want to use real sources for tests,
	// and we cannot reliably wait for a source ingestion to complete before trying to get the sub-sources.
	sourceId, err := createSourceSchemaAndGetSourceId(ctx, client, accountId)
	if err != nil {
		t.Fatalf("Failed to initialize source schema and get source id: %v", err)
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
			return deleteSource(ctx, client, sourceId)
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

// Helpers

func createSourceSchemaAndGetSourceId(ctx context.Context, client *sifflet.ClientWithResponses, accountId string) (types.UUID, error) {
	// Create the source v1 (source schema)
	subSourceName := providertests.RandomName()
	subSourceDescription := "Created by Terraform provider tests"

	awsRegion := "eu-west-1"
	roleArn := "arn:aws:iam::123456789012:role/role"
	// We use a AWS Quicksight source because it does not require secrets, so it is easier to setup.
	dto := sifflet.PublicQuicksightParametersDto{
		Type:      sifflet.PublicQuicksightParametersDtoTypeQUICKSIGHT,
		AccountId: &accountId,
		AwsRegion: &awsRegion,
		RoleArn:   &roleArn,
	}
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	err := parametersDto.FromPublicQuicksightParametersDto(dto)
	if err != nil {
		return types.UUID{}, fmt.Errorf("Failed to create parameters: %v", err)
	}
	source := sifflet.PublicCreateSourceDto{
		Name:        subSourceName,
		Description: &subSourceDescription,
		Parameters:  parametersDto,
	}
	sourceResponse, err := client.PublicCreateSourceWithResponse(ctx, source)
	if err != nil {
		return types.UUID{}, fmt.Errorf("Failed to create source: %v", err)
	}
	if sourceResponse.StatusCode() != http.StatusCreated {
		return types.UUID{}, fmt.Errorf("Failed to create source: status code %d", sourceResponse.StatusCode())
	}

	// Get the source v2 id corresponding to the created sub-source
	// The source v2 getAll endpoint does not allow filtering by name, so we need to iterate over all sources.
	var sourceId types.UUID
	sourcesResponse, err := client.PublicGetSourcesV2WithResponse(ctx)
	if err != nil {
		return types.UUID{}, fmt.Errorf("Failed to get sources: %v", err)
	}
	if sourcesResponse.StatusCode() != http.StatusOK {
		return types.UUID{}, fmt.Errorf("Failed to get sources: status code %d", sourcesResponse.StatusCode())
	}
	sources := sourcesResponse.JSON200.Data
	if len(sources) == 0 {
		return types.UUID{}, fmt.Errorf("No sources found")
	}
	for _, source := range sources {
		var typedSource sifflet.SiffletPublicGetSourceV2Dto
		err := typedSource.FromPublicPageDtoPublicGetSourceV2DtoDataItem(source)
		if err != nil {
			return types.UUID{}, fmt.Errorf("Failed to get source: %v", err)
		}
		sourceItem, err := typedSource.GetSourceDto()
		if err != nil {
			return types.UUID{}, fmt.Errorf("Error reading source: %s", err)
		}
		if strings.Contains(sourceItem.GetName(), providertests.SessionPrefix()) {
			sourceId = sourceItem.GetId()
			break
		}
	}
	if sourceId == (types.UUID{}) {
		return types.UUID{}, fmt.Errorf("No source found with name containing %s", providertests.SessionPrefix())
	}

	return sourceId, nil
}

func deleteSource(ctx context.Context, client *sifflet.ClientWithResponses, sourceId types.UUID) error {
	sourceResponse, err := client.PublicDeleteSourceV2WithResponse(ctx, sourceId)
	if err != nil {
		return fmt.Errorf("Failed to delete source: %v", err)
	}
	if sourceResponse.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("Failed to delete source: status code %d", sourceResponse.StatusCode())
	}
	return nil
}
