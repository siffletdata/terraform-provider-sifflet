package domain

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &domainDataSource{}
	_ datasource.DataSourceWithConfigure = &domainDataSource{}
)

func newDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

type domainDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func DomainDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Read a Sifflet domain by its ID. A domain represents a subset of assets. Domains are used to provide a view of specific business areas, and provide different access to different teams.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the domain.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the domain.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the domain.",
				Computed:    true,
			},
		},
	}
}

func (d *domainDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DomainDataSourceSchema(ctx)
}

type domainDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var data domainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idStr := data.Id.ValueString()
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse domain ID",
			err.Error(),
		)
		return
	}

	var domainResponse *sifflet.PublicGetDomainResponse

	domainResponse, err = d.client.PublicGetDomainWithResponse(ctx, id)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read asset",
			err.Error(),
		)
		return
	}

	if domainResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read asset",
			domainResponse.StatusCode(), domainResponse.Body,
		)
		return
	}

	data.Name = types.StringValue(domainResponse.JSON200.Name)
	data.Description = types.StringPointerValue(domainResponse.JSON200.Description)

	diags := resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
