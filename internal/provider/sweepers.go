package provider

import (
	"context"
	"errors"
	"os"
	"terraform-provider-sifflet/internal/apiclients"
	"terraform-provider-sifflet/internal/client"
)

// ClientForSweepers creates a new Sifflet API client to be used to write sweepers (functions that delete dangling acceptance test resources).
func ClientForSweepers(ctx context.Context) (*client.ClientWithResponses, error) {
	token := os.Getenv("SIFFLET_TOKEN")
	host := os.Getenv("SIFFLET_HOST")
	clients, diag := apiclients.MakeHttpClients(
		ctx, token, host, "sweeper", "sweeper",
	)
	if diag != nil {
		return nil, errors.New(diag.Summary())
	}
	return clients.Client, nil
}
