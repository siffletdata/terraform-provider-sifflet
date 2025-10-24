package asset

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSource              = &assetsDataSource{}
	_ datasource.DataSourceWithConfigure = &assetsDataSource{}
)

func newAssetsDataSource() datasource.DataSource {
	return &assetsDataSource{}
}

type assetsDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *assetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *assetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets"
}

func AssetsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Return assets matching search criteria.",
		Attributes: map[string]schema.Attribute{
			"max_results": schema.Int32Attribute{
				Description: "Maximum number of results to return. Default is 1000.",
				Optional:    true,
			},
			"filter": schema.SingleNestedAttribute{
				Description: "Search criteria",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"text_search": schema.StringAttribute{
						Description: "Return assets whose name match this attribute",
						Optional:    true,
					},
					"type_categories": schema.ListAttribute{
						Description: "List of asset type categories to filter on. Valid values are TABLE_AND_VIEW, PIPELINE, DASHBOARD, ML_MODEL. For filtering declared assets with custom types, you can use the format declared-asset_{custom sub type}. For example: declared-asset_Storage",
						ElementType: types.StringType,
						Optional:    true,
					},
					"tags": schema.ListNestedAttribute{
						Description: "List of tags to filter assets by. Tags can be identified by either ID or name. If a name is provided, optionally a kind can be provided to disambiguate tags of different types sharing the same name.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Tag ID",
									Optional:    true,
								},
								"name": schema.StringAttribute{
									Description: "Tag name",
									Optional:    true,
								},
								"kind": schema.StringAttribute{
									Description: "Tag kind (such as 'Tag' or 'Classification')",
									Optional:    true,
									Validators: []validator.String{
										stringvalidator.OneOf("Tag", "Classification"),
										stringvalidator.ConflictsWith(
											path.MatchRelative().AtParent().AtName("id"),
										),
									},
								},
							},
						},
					},
				},
			},
			"results": schema.ListNestedAttribute{
				Description: "List of assets returned by the search",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Asset ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Asset name",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Asset type. This is the specific type of the asset, not the broader type category used in the filter. For example, an asset in type category TABLE_AND_VIEW can have the type TABLE.",
							Computed:    true,
						},
						"urn": schema.StringAttribute{
							Description: "Internal Sifflet identifier for the asset",
							Computed:    true,
						},
						"uri": schema.StringAttribute{
							Description: "URI string identifying the asset. More about URIs here: https://docs.siffletdata.com/docs/uris",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Asset description",
							Computed:    true,
						},
						"tags": schema.ListNestedAttribute{
							Description: "List of tags associated with this source",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "Tag ID",
										Computed:    true,
									},
									"name": schema.StringAttribute{
										Description: "Tag name",
										Computed:    true,
									},
									"kind": schema.StringAttribute{
										Description: "Tag kind (such as 'Tag' or 'Classification')",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *assetsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AssetsDataSourceSchema(ctx)
}

type AssetsDataSourceModel struct {
	MaxResults types.Int32  `tfsdk:"max_results"`
	Filter     types.Object `tfsdk:"filter"`
	Results    types.List   `tfsdk:"results"`
}

type FilterModel struct {
	Tags           types.List   `tfsdk:"tags"`
	TextSearch     types.String `tfsdk:"text_search"`
	TypeCategories types.List   `tfsdk:"type_categories"`
}

func (m FilterModel) ToDto(ctx context.Context) (sifflet.PublicAssetFilterDto, diag.Diagnostics) {
	var typeCategories []string

	diags := m.TypeCategories.ElementsAs(ctx, &typeCategories, false)
	if diags.HasError() {
		return sifflet.PublicAssetFilterDto{}, diags
	}

	var tags []tagModel
	diags = m.Tags.ElementsAs(ctx, &tags, false)
	if diags.HasError() {
		return sifflet.PublicAssetFilterDto{}, diags
	}

	tagsDto, diags := tfutils.MapWithDiagnostics(tags, func(tag tagModel) (sifflet.PublicTagReferenceDto, diag.Diagnostics) {
		return tag.ToDto()
	})
	if diags.HasError() {
		return sifflet.PublicAssetFilterDto{}, diags
	}

	return sifflet.PublicAssetFilterDto{
		TextSearch: m.TextSearch.ValueStringPointer(),
		AssetType:  &typeCategories,
		Tags:       &tagsDto,
	}, diag.Diagnostics{}
}

func (d *assetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var data AssetsDataSourceModel
	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var filterModel FilterModel
	diags = data.Filter.As(ctx, &filterModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	filterDto, diags := filterModel.ToDto(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var page int32 = 0
	var itemsPerPage int32 = 100
	remainingResults := data.MaxResults.ValueInt32()
	if remainingResults == 0 {
		// Set a default
		remainingResults = 1000
	}
	results := make([]assetModel, 0)

	for ; ; page++ {
		if remainingResults <= itemsPerPage {
			itemsPerPage = remainingResults
		}

		paginationDto := sifflet.PublicAssetPaginationDto{
			ItemsPerPage: &itemsPerPage,
			Page:         &page,
		}
		requestDto := sifflet.PublicAssetSearchCriteriaDto{
			Filter:     &filterDto,
			Pagination: &paginationDto,
		}
		searchResponse, err := d.client.PublicGetAssetsWithResponse(ctx, requestDto)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read assets", err.Error())
			return
		}
		if searchResponse.StatusCode() != http.StatusOK {
			sifflet.HandleHttpErrorAsProblem(
				ctx, &resp.Diagnostics, "Unable to read assets", searchResponse.StatusCode(), searchResponse.Body,
			)
			return
		}

		responseDto := *searchResponse.JSON200
		if len(responseDto.Data) == 0 {
			break
		}
		remainingResults -= int32(len(responseDto.Data)) // nolint: gosec
		if remainingResults <= 0 {
			break
		}
		for _, data := range responseDto.Data {
			var assetModel assetModel
			diags := assetModel.FromListDto(ctx, data)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}
			results = append(results, assetModel)
		}
	}

	data.Results, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: assetModel{}.AttributeTypes()}, results)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}
