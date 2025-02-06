package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"terraform-provider-sifflet/internal/apiclients"
	"terraform-provider-sifflet/internal/provider/credentials"
	sifflet_datasource "terraform-provider-sifflet/internal/provider/datasource"
	"terraform-provider-sifflet/internal/provider/source"
	"terraform-provider-sifflet/internal/provider/tag"
	"terraform-provider-sifflet/internal/provider/user"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &siffletProvider{}
)

type siffletProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &siffletProvider{
			version: version,
		}
	}
}

// siffletProvider is the provider implementation.
type siffletProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *siffletProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sifflet"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *siffletProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "Sifflet API host, such as `https://yourinstance.siffletdata.com/api`. If not set, the provider will use the SIFFLET_HOST environment variable.",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Sifflet API token. If not set, the provider will use the SIFFLET_TOKEN environment variable. We recommend not setting this value directly in the configuration, use the environment variable instead.",
			},
		},
	}
}

// Configure prepares a sifflet API client for data sources and resources.
func (p *siffletProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config siffletProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Sifflet API Host",
			"The provider cannot create the Sifflet API client as there is an unknown configuration value for the Sifflet API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration (under the `host` attribute of the provider block), or use the SIFFLET_HOST environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Sifflet API token",
			"The provider cannot create the Sifflet API client as there is an unknown configuration value for the Sifflet API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration (under the `token` attribute of the provider block), or use the SIFFLET_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("SIFFLET_HOST")
	token := os.Getenv("SIFFLET_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Sifflet API Host",
			"The provider cannot create the Sifflet API client as there is a missing or empty value for the Sifflet API host. "+
				"Set the host value in the configuration or use the SIFFLET_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Sifflet API Username",
			"The provider cannot create the Sifflet API client as there is a missing or empty value for the Sifflet API token. "+
				"Set the token value in the provider configuration or use the SIFFLET_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tfVersion := req.TerraformVersion
	providerVersion := p.version
	httpClients, diag := apiclients.MakeHttpClients(ctx, token, host, tfVersion, providerVersion)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	// Make the Sifflet clients available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = httpClients
	resp.ResourceData = httpClients

	// Check that the provided URL is valid by making a request
	// to the Sifflet API.
	apiHealthCheck(httpClients.HttpClient, host, resp)
}

func apiHealthCheck(httpClient *http.Client, host string, resp *provider.ConfigureResponse) {
	queryUrl, err := url.JoinPath(host, "/actuator/health")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate Sifflet API Host",
			fmt.Sprintf("Got error when building the health check URL from the configured Sifflet API host %s: %s ", host, err.Error()),
		)
		return
	}
	res, err := httpClient.Get(queryUrl)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate Sifflet API Host",
			fmt.Sprintf("Got error when attempting to perform a health check on the configured Sifflet API host %s: %s ", host, err.Error()),
		)
		return
	}
	if res.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Unable to validate Sifflet API Host",
			fmt.Sprintf("Got an unexpected status code when attempting to perform a health check on the configured Sifflet API host %s: expected 200, got %d ", host, res.StatusCode),
		)
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *siffletProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return slices.Concat(
		credentials.DataSources(),
		sifflet_datasource.DataSources(),
		source.DataSources(),
		tag.DataSources(),
		user.DataSources(),
	)
}

// Resources defines the resources implemented in the provider.
func (p *siffletProvider) Resources(_ context.Context) []func() resource.Resource {
	return slices.Concat(
		credentials.Resources(),
		sifflet_datasource.Resources(),
		source.Resources(),
		tag.Resources(),
		user.Resources(),
	)
}

var (
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sifflet": providerserver.NewProtocol6WithError(New("test")()),
	}
)
