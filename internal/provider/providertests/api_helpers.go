package providertests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
)

// ClientForTests creates a new Sifflet API client to be used to write tests.
func ClientForTests(ctx context.Context) (*sifflet.ClientWithResponses, error) {
	token := os.Getenv("SIFFLET_TOKEN")
	host := os.Getenv("SIFFLET_HOST")
	clients, diag := apiclients.MakeHttpClients(
		ctx, token, host, "test", "test",
	)
	if diag != nil {
		return nil, errors.New(diag.Summary())
	}
	return clients.Client, nil
}

func CreateDeclaredAssets(ctx context.Context, client *sifflet.ClientWithResponses, workspaceName string, assets *[]sifflet.PublicDeclarativeAssetDto) error {
	dryRun := false

	payload := sifflet.PublicDeclarativePayloadDto{
		Assets:    assets,
		Workspace: workspaceName,
	}
	params := sifflet.PublicSyncAssetsParams{
		DryRun: &dryRun,
	}

	response, err := client.PublicSyncAssetsWithResponse(ctx, &params, payload)
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to sync assets: status code %d. Details: %s", response.StatusCode(), response.Body)
	}
	return nil
}

func DeleteDeclaredAssets(ctx context.Context, client *sifflet.ClientWithResponses, workspaceName string) error {
	dryRun := false

	response, err := client.PublicDeleteWorkspaceWithResponse(ctx, workspaceName, &sifflet.PublicDeleteWorkspaceParams{DryRun: &dryRun})
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete workspace: status code %d. Details: %s", response.StatusCode(), response.Body)
	}
	return nil
}
