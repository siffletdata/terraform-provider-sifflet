package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &credentialDataSource{}
	_ datasource.DataSourceWithConfigure = &credentialDataSource{}
)

func NewCredentialDataSource() datasource.DataSource {
	return &credentialDataSource{}
}

type credentialDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *credentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*httpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *httpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = clients.Client
}

func (d *credentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func CredentialDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Read a Sifflet credential. This data source doesn't return the credential value.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the credential.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the credential",
				Computed:    true,
			},
		},
	}
}

func (d *credentialDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CredentialDataSourceSchema(ctx)
}

type CredentialDataSourceDto struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *credentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CredentialDataSourceDto
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	maxAttempts := 20
	var credentialResponse *sifflet.PublicGetCredentialResponse
	var err error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		credentialResponse, err = d.client.PublicGetCredentialWithResponse(ctx, name)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read credential",
				err.Error(),
			)
			return
		}
		if credentialResponse.StatusCode() == http.StatusNotFound {
			// Retry a few times, as there's a delay in the API (eventual consistency)
			if attempt < maxAttempts {
				time.Sleep(200 * time.Millisecond)
				continue
			}
		} else {
			break
		}
	}

	if credentialResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read credential",
			credentialResponse.StatusCode(), credentialResponse.Body,
		)
		return
	}

	data.Name = types.StringValue(credentialResponse.JSON200.Name)
	data.Description = types.StringPointerValue(credentialResponse.JSON200.Description)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
