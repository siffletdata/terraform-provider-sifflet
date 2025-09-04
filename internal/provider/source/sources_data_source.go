package source

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/source/parameters"
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
	_ datasource.DataSource              = &sourcesDataSource{}
	_ datasource.DataSourceWithConfigure = &sourcesDataSource{}
)

func newSourcesDataSource() datasource.DataSource {
	return &sourcesDataSource{}
}

type sourcesDataSource struct {
	client *sifflet.ClientWithResponses
}

func (d *sourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sources"
}

func SourcesDataSourceSchema(ctx context.Context) schema.Schema {
	paramsSchema := parameters.ParametersModel{}.TerraformSchema()
	paramsSchema.Computed = true
	paramsSchema.Required = false
	return schema.Schema{
		Description: "Return sources matching search criteria.",
		Attributes: map[string]schema.Attribute{
			"max_results": schema.Int32Attribute{
				Description: "Maximum number of results to return. Default is 1000.",
				Optional:    true,
			},
			"filter": schema.SingleNestedAttribute{
				Description: "Search criteria",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"tags": schema.ListNestedAttribute{
						Description: "List of tags to filter sources by. Tags can be identified by either ID or name. If a name is provided, optionally a kind can be provided to disambiguate tags of different types sharing the same name.",
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
					"text_search": schema.StringAttribute{
						Description: "Return sources whose name match this attribute",
						Optional:    true,
					},
					"types": schema.ListAttribute{
						Description: "List of source types to filter sources by",
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"results": schema.ListNestedAttribute{
				Description: "List of sources returned by the search",
				Computed:    true,
				// This currently duplicates the schema of the "sifflet_source" resource, consider merging them.
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Source ID",
							Computed:    true,
						},
						"credentials": schema.StringAttribute{
							Description: "Name of the credentials used to connect to the source",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Source name",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Source description",
							Computed:    true,
						},
						"parameters": paramsSchema,
						"schedule": schema.StringAttribute{
							Description: "When this source is scheduled to refresh (cron expression). Will be null if not scheduled.",
							Computed:    true,
						},
						"timezone": schema.StringAttribute{
							Description: "Timezone for this source",
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

func (d *sourcesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SourcesDataSourceSchema(ctx)
}

type SourcesDataSourceModel struct {
	MaxResults types.Int32  `tfsdk:"max_results"`
	Filter     types.Object `tfsdk:"filter"`
	Results    types.List   `tfsdk:"results"`
}

type FilterModel struct {
	Tags       types.List   `tfsdk:"tags"`
	TextSearch types.String `tfsdk:"text_search"`
	Types      types.List   `tfsdk:"types"`
}

func (m FilterModel) ToDto(ctx context.Context) (sifflet.PublicSourceFilterDto, diag.Diagnostics) {
	var types []sifflet.PublicSourceFilterDtoTypes

	diags := m.Types.ElementsAs(ctx, &types, false)
	if diags.HasError() {
		return sifflet.PublicSourceFilterDto{}, diags
	}

	var tags []tagModel
	diags = m.Tags.ElementsAs(ctx, &tags, false)
	if diags.HasError() {
		return sifflet.PublicSourceFilterDto{}, diags
	}

	tagsDto, diags := tfutils.MapWithDiagnostics(tags, func(tag tagModel) (sifflet.PublicTagReferenceDto, diag.Diagnostics) {
		return tag.ToDto()
	})
	if diags.HasError() {
		return sifflet.PublicSourceFilterDto{}, diags
	}

	return sifflet.PublicSourceFilterDto{
		TextSearch: m.TextSearch.ValueStringPointer(),
		Types:      &types,
		Tags:       &tagsDto,
	}, diag.Diagnostics{}
}

func (d *sourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var data SourcesDataSourceModel
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
	results := make([]baseSourceModel, 0)

	for ; ; page++ {
		if remainingResults <= itemsPerPage {
			itemsPerPage = remainingResults
		}

		paginationDto := sifflet.PublicSourcePaginationDto{
			ItemsPerPage: &itemsPerPage,
			Page:         &page,
		}
		requestDto := sifflet.PublicSourceSearchCriteriaDto{
			Filter:     &filterDto,
			Pagination: &paginationDto,
		}
		searchResponse, err := d.client.PublicGetSourcesWithResponse(ctx, requestDto)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read sources", err.Error())
			return
		}
		if searchResponse.StatusCode() != http.StatusOK {
			sifflet.HandleHttpErrorAsProblem(
				ctx, &resp.Diagnostics, "Unable to read sources", searchResponse.StatusCode(), searchResponse.Body,
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
			var baseSourceModel baseSourceModel
			diags := baseSourceModel.FromDto(ctx, data)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}
			results = append(results, baseSourceModel)
		}
	}

	data.Results, diags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: baseSourceModel{}.AttributeTypes()}, results)
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
