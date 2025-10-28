package asset

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &assetDataSource{}
	_ datasource.DataSourceWithConfigure = &assetDataSource{}
)

func newAssetDataSource() datasource.DataSource {
	return &assetDataSource{}
}

type assetDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *assetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *assetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func AssetDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Read a Sifflet asset by its URI.",
		Attributes: map[string]schema.Attribute{
			"uri": schema.StringAttribute{
				Description: "URI string identifying the asset. More about URIs here: https://docs.siffletdata.com/docs/uris.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Id of the asset.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the asset.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the asset.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the asset.",
				Computed:    true,
			},
			"tags": schema.ListNestedAttribute{
				Description: "List of tags associated with this source.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Tag ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Tag name.",
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Description: "Tag kind (such as 'Tag' or 'Classification').",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *assetDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AssetDataSourceSchema(ctx)
}

func (d *assetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var data assetModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uri := data.Uri.ValueString()

	maxAttempts := 20
	var assetResponse *sifflet.PublicGetAssetResponse
	var err error

	request := sifflet.PublicGetAssetRequestDto{Uri: uri}

	for attempt := range maxAttempts {
		assetResponse, err = d.client.PublicGetAssetWithResponse(ctx, request)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read asset",
				err.Error(),
			)
			return
		}
		if assetResponse.StatusCode() == http.StatusNotFound {
			// Retry a few times, as there's a delay in the API (eventual consistency)
			if attempt < maxAttempts {
				time.Sleep(200 * time.Millisecond)
				continue
			}
		} else {
			break
		}
	}

	if assetResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read asset",
			assetResponse.StatusCode(), assetResponse.Body,
		)
		return
	}

	diags := data.FromDto(ctx, *assetResponse.JSON200)
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
