package credentials

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &credentialsDataSource{}
	_ datasource.DataSourceWithConfigure = &credentialsDataSource{}
)

func newCredentialDataSource() datasource.DataSource {
	return &credentialsDataSource{}
}

type credentialsDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *credentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*apiclients.HttpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HttpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = clients.Client
}

func (d *credentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credentials"
}

func CredentialDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Read Sifflet credentials. This data source doesn't return the credentials value.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the credentials.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the credentials",
				Computed:    true,
			},
		},
	}
}

func (d *credentialsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CredentialDataSourceSchema(ctx)
}

type CredentialsDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *credentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CredentialsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	maxAttempts := 20
	var credentialsResponse *sifflet.PublicGetCredentialsResponse
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Conflict in the OpenAPI schema between operation IDs for "get credentials" and "list credentials",
		// hence the strange operation name.
		credentialsResponse, err = d.client.PublicGetCredentialsWithResponse(ctx, name)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read credentials",
				err.Error(),
			)
			return
		}
		if credentialsResponse.StatusCode() == http.StatusNotFound {
			// Retry a few times, as there's a delay in the API (eventual consistency)
			if attempt < maxAttempts {
				time.Sleep(200 * time.Millisecond)
				continue
			}
		} else {
			break
		}
	}

	if credentialsResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read credentials",
			credentialsResponse.StatusCode(), credentialsResponse.Body,
		)
		return
	}

	data.Name = types.StringValue(credentialsResponse.JSON200.Name)
	data.Description = types.StringPointerValue(credentialsResponse.JSON200.Description)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
