package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	alphasifflet "terraform-provider-sifflet/internal/alphaclient"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
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
		// TODO: add validators and descriptions to the schema
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
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
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SIFFLET_HOST environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Sifflet API token",
			"The provider cannot create the Sifflet API client as there is an unknown configuration value for the Sifflet API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SIFFLET_TOKEN environment variable.",
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
				"Set the token value in the configuration or use the SIFFLET_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(token)
	if bearerTokenProviderErr != nil {
		panic(bearerTokenProviderErr)
	}

	// Exhaustive list of some defaults you can use to initialize a Client.
	// If you need to override the underlying httpClient, you can use the option
	//
	// WithHTTPClient(httpClient *http.Client)
	//

	// Create a new Sifflet alphaclient using the configuration values
	alphaclient, err := alphasifflet.NewClient(host, alphasifflet.WithRequestEditorFn(bearerTokenProvider.Intercept))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Sifflet API Client",
			"An unexpected error occurred when creating the Sifflet API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Sifflet Client Error: "+err.Error(),
		)
		return
	}

	client, err := sifflet.NewClientWithResponses(host, sifflet.WithRequestEditorFn(bearerTokenProvider.Intercept))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Sifflet API Client",
			"An unexpected error occurred when creating the Sifflet API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Sifflet Client Error: "+err.Error(),
		)
		return
	}

	httpClients := &httpClients{
		AlphaClient: alphaclient,
		Client:      client,
	}

	// Make the Sifflet clients available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = httpClients
	resp.ResourceData = httpClients
}

type httpClients struct {
	AlphaClient *alphasifflet.Client
	Client      *sifflet.ClientWithResponses
}

// DataSources defines the data sources implemented in the provider.
func (p *siffletProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDatasourcesDataSource,
		NewTagDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *siffletProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDataSourceResource,
		NewTagResource,
	}
}
