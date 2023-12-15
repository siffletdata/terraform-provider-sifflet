package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &datasourcesDataSource{}
	_ datasource.DataSourceWithConfigure = &datasourcesDataSource{}
)

type DatasourceSearchDto struct {
	CatalogFilters []CatalogFilterDto                        `tfsdk:"catalog_filters"`
	SearchRules    SearchCollectionDatasourceCatalogAssetDto `tfsdk:"search_rules"`
}

type CatalogFilterDto struct {
	// Children *[]FilterElementDto `tfsdk:"children"`
	Id    *string `tfsdk:"id"`
	Name  *string `tfsdk:"name"`
	Query *string `tfsdk:"query"`
}

type DatasourceCatalogAssetDtoEntityType string

type DatasourceCatalogAssetDto struct {
	CreatedBy        *string                             `tfsdk:"created_by"`
	CreatedDate      *int64                              `tfsdk:"created_date"`
	CronExpression   *string                             `tfsdk:"cron_expression"`
	EntityType       DatasourceCatalogAssetDtoEntityType `tfsdk:"entity_type"`
	Id               *string                             `tfsdk:"id"`
	LastModifiedDate *int64                              `tfsdk:"last_modified_date"`
	// LastWeekStatuses *[]LastIngestionStatusDto           `tfsdk:"lastWeekStatuses"`
	ModifiedBy    *string `tfsdk:"modified_by"`
	Name          string  `tfsdk:"name"`
	NextExecution *int64  `tfsdk:"next_execution"`
	// Params           DatasourceCatalogAssetDto_Params    `tfsdk:"params"`
	// Tags             *[]TagDto                           `tfsdk:"tags"`
	Type string `tfsdk:"type"`
}

type SearchCollectionDatasourceCatalogAssetDto struct {
	Data          *[]DatasourceCatalogAssetDto `tfsdk:"data"`
	TotalElements *int64                       `tfsdk:"total_elements"`
}

func NewDatasourcesDataSource() datasource.DataSource {
	return &datasourcesDataSource{}
}

type datasourcesDataSource struct {
	client *sifflet.Client
}

func (d *datasourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sifflet.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sifflet.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *datasourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasources"
}

func (d *datasourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"catalog_filters": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// "children": schema.ListNestedAttribute{
						// 	NestedObject: schema.NestedAttributeObject{
						// 		Attributes: map[string]schema.Attribute{
						// 			"id": schema.Int64Attribute{
						// 				Computed: true,
						// 			},
						// 			"name": schema.StringAttribute{
						// 				Computed: true,
						// 			},
						// 			"results": schema.Int64Attribute{
						// 				Computed: true,
						// 			},
						// 		},
						// 	},
						// 	Computed: true,
						// },
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"query": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"search_rules": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"data": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"created_by": schema.StringAttribute{
									Computed: true,
								},
								"created_date": schema.Int64Attribute{
									Computed: true,
								},
								"cron_expression": schema.StringAttribute{
									Computed: true,
								},
								"entity_type": schema.StringAttribute{
									Computed: true,
								},
								"id": schema.StringAttribute{
									Computed: true,
								},
								"last_modified_date": schema.Int64Attribute{
									Computed: true,
								},
								"modified_by": schema.StringAttribute{
									Computed: true,
								},
								"name": schema.StringAttribute{
									Computed: true,
								},
								"next_execution": schema.Int64Attribute{
									Computed: true,
								},
								"type": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						Computed: true,
					},
					"total_elements": schema.Int64Attribute{
						Computed: true,
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *datasourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DatasourceSearchDto

	ItemsPerPage := int32(-1)

	params := sifflet.GetAllDatasourceParams{
		ItemsPerPage: &ItemsPerPage,
	}

	itemResponse, err := d.client.GetAllDatasource(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read HashiCups Coffees",
			err.Error(),
		)
		return
	}
	resBody, _ := io.ReadAll(itemResponse.Body)
	tflog.Debug(ctx, "test1 "+string(resBody))

	var result sifflet.DatasourceSearchDto
	if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
		resp.Diagnostics.AddError(
			"Can not unmarshal JSON",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("TotalElements: %d\n", *result.SearchDatasources.TotalElements))

	state.SearchRules.TotalElements = result.SearchDatasources.TotalElements

	if state.SearchRules.Data == nil {
		state.SearchRules.Data = &[]DatasourceCatalogAssetDto{}
	}

	for _, data := range *result.SearchDatasources.Data {
		idString := data.Id.String()
		yEntityType := DatasourceCatalogAssetDtoEntityType(data.EntityType)
		*state.SearchRules.Data = append(*state.SearchRules.Data, DatasourceCatalogAssetDto{
			CreatedBy:        data.CreatedBy,
			CreatedDate:      data.CreatedDate,
			CronExpression:   data.CronExpression,
			LastModifiedDate: data.LastModifiedDate,
			ModifiedBy:       data.ModifiedBy,
			Name:             data.Name,
			NextExecution:    data.NextExecution,
			Type:             data.Type,
			EntityType:       yEntityType,
			Id:               &idString,
		})
	}

	for _, filters := range result.CatalogFilters {
		state.CatalogFilters = append(state.CatalogFilters, CatalogFilterDto{
			Name:  filters.Name,
			Id:    filters.Id,
			Query: filters.Query,
		})
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
