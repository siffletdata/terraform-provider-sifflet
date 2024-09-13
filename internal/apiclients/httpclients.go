package apiclients

import (
	"net/http"
	alphasifflet "terraform-provider-sifflet/internal/alphaclient"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfhttp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

type HttpClients struct {
	AlphaClient *alphasifflet.ClientWithResponses
	Client      *sifflet.ClientWithResponses
	HttpClient  *http.Client
}

func MakeHttpClients(token string, host string) (*HttpClients, diag.Diagnostic) {
	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(token)
	if bearerTokenProviderErr != nil {
		panic(bearerTokenProviderErr)
	}

	alphaclient, err := alphasifflet.NewClientWithResponses(host, alphasifflet.WithRequestEditorFn(bearerTokenProvider.Intercept))
	if err != nil {
		return nil, diag.NewErrorDiagnostic(
			"Unable to Create Sifflet API Client",
			"An unexpected error occurred when creating the Sifflet API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Sifflet Client Error: "+err.Error(),
		)
	}

	httpClient := tfhttp.NewTerraformHttpClient()

	client, err := sifflet.NewClientWithResponses(
		host,
		sifflet.WithRequestEditorFn(bearerTokenProvider.Intercept),
		sifflet.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic(
			"Unable to Create Sifflet API Client",
			"An unexpected error occurred when creating the Sifflet API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Sifflet Client Error: "+err.Error(),
		)
	}

	return &HttpClients{
		AlphaClient: alphaclient,
		Client:      client,
		HttpClient:  httpClient,
	}, nil
}
