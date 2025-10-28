package source_v2

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &sourceV2SchemasDataSource{}
	_ datasource.DataSourceWithConfigure = &sourceV2SchemasDataSource{}
)

func newSourceV2SchemasDataSource() datasource.DataSource {
	return &sourceV2SchemasDataSource{}
}

type sourceV2SchemasDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *sourceV2SchemasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sourceV2SchemasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_v2_schemas"
}

func SourceV2SchemasDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: `Get all schemas for a given source.

"Schemas" here is used broadly to refer to all underlying schemas/workspaces/databases/instances contained in a Sifflet source.

This will return the URIs of the schemas for the source. Read more about URIs here: https://docs.siffletdata.com/docs/uri.

For example:
- For an Airflow source, this will return the URIs of the Airflow instances contained in the source: https://docs.siffletdata.com/docs/airflow-uri-format.
- For an Athena source, this will return the URIs of the databases contained in the source: https://docs.siffletdata.com/docs/athena-uri-format.
- For a MySQL source, this will return the URIs of the schemas contained in the source: https://docs.siffletdata.com/docs/mysql-uri-format.

~> Note that the schemas are not immediately available after the source is created. You need to wait for the initial source ingestion to finish before the schemas are available. As such, we do not recomment using this data source in the same module as the one that creates the corresponding Sifflet source.`,
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Description: "Id of the source.",
				Required:    true,
			},
			"schemas": schema.ListNestedAttribute{
				Description: "List of schemas/workspaces/databases/instances contained in the source.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uri": schema.StringAttribute{
							Description: "URI of the schema.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *sourceV2SchemasDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SourceV2SchemasDataSourceSchema(ctx)
}

type sourceV2SchemasModel struct {
	SourceId types.String `tfsdk:"source_id"`
	Schemas  types.List   `tfsdk:"schemas"`
}

type schemaModel struct {
	Uri types.String `tfsdk:"uri"`
}

func (m schemaModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"uri": types.StringType,
	}
}

func (d *sourceV2SchemasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var data sourceV2SchemasModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceId, err := uuid.Parse(data.SourceId.ValueString())
	resp.Diagnostics.Append(tfutils.ErrToDiags("Could not parse ID as UUID", err)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assetResponse, err := d.client.PublicGetSourceSchemaListWithResponse(ctx, sourceId)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read source schemas",
			err.Error(),
		)
		return
	}

	if assetResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read source schemas",
			assetResponse.StatusCode(), assetResponse.Body,
		)
		return
	}

	results := make([]schemaModel, 0)
	for _, schema := range assetResponse.JSON200.Schemas {
		results = append(results, schemaModel{Uri: types.StringValue(schema.Uri)})
	}

	var diags diag.Diagnostics
	data.Schemas, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: schemaModel{}.AttributeTypes()}, results)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
