package datasource_datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TimeZoneDto struct {
	TimeZone  types.String `tfsdk:"timezone"`
	UtcOffset types.String `tfsdk:"utc_offset"`
}

type BigQueryParams struct {
	Type             types.String `tfsdk:"type"`
	BillingProjectID *string      `tfsdk:"billing_project_id"`
	DatasetID        *string      `tfsdk:"dataset_id"`
	ProjectID        *string      `tfsdk:"project_id"`
	TimezoneData     *TimeZoneDto `tfsdk:"timezone_data"`
}

type DBTParams struct {
	Type         types.String `tfsdk:"type"`
	ProjectName  *string      `tfsdk:"project_name"`
	Target       *string      `tfsdk:"target"`
	TimezoneData *TimeZoneDto `tfsdk:"timezone_data"`
}

type CreateDatasourceDto struct {
	ID             types.String    `tfsdk:"id"`
	Name           *string         `tfsdk:"name"`
	CronExpression *string         `tfsdk:"cron_expression"`
	Type           types.String    `tfsdk:"type"`
	SecretID       *string         `tfsdk:"secret_id"`
	BigQuery       *BigQueryParams `tfsdk:"bigquery"`
	DBT            *DBTParams      `tfsdk:"dbt"`
	CreatedBy      types.String    `tfsdk:"created_by"`
	CreatedDate    types.String    `tfsdk:"created_date"`
	ModifiedBy     types.String    `tfsdk:"modified_by"`
}

type ErrorMessage struct {
	Title  string `json:"title"`
	Status int64  `json:"status"`
	Detail string `json:"detail"`
}

func DatasourceDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cron_expression": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"secret_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_date": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bigquery": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"billing_project_id": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"dataset_id": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"project_id": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"timezone_data": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"timezone":   types.StringType,
									"utc_offset": types.StringType,
								},
								map[string]attr.Value{
									"timezone":   types.StringValue("UTC"),
									"utc_offset": types.StringValue("(UTC+00:00)"),
								},
							),
						),
						Attributes: map[string]schema.Attribute{
							"timezone": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"utc_offset": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"dbt": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"project_name": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"target": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"timezone_data": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"timezone":   types.StringType,
									"utc_offset": types.StringType,
								},
								map[string]attr.Value{
									"timezone":   types.StringValue("UTC"),
									"utc_offset": types.StringValue("(UTC+00:00)"),
								},
							),
						),
						Attributes: map[string]schema.Attribute{
							"timezone": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"utc_offset": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
		},
	}
}

type DatasourceCatalogAssetDtoEntityType string

type DatasourceCatalogAssetDto struct {
	CreatedBy        *string                             `tfsdk:"created_by"`
	CreatedDate      *int64                              `tfsdk:"created_date"`
	CronExpression   *string                             `tfsdk:"cron_expression"`
	EntityType       DatasourceCatalogAssetDtoEntityType `tfsdk:"entity_type"`
	Id               *string                             `tfsdk:"id"`
	LastModifiedDate *int64                              `tfsdk:"last_modified_date"`
	ModifiedBy       *string                             `tfsdk:"modified_by"`
	Name             string                              `tfsdk:"name"`
	NextExecution    *int64                              `tfsdk:"next_execution"`
	BigQuery         *BigQueryParams                     `tfsdk:"bigquery"`
	DBT              *DBTParams                          `tfsdk:"dbt"`
	// Tags             *[]TagDto                           `tfsdk:"tags"`
	Type string `tfsdk:"type"`
}

type SearchCollectionDatasourceCatalogAssetDto struct {
	Data          *[]DatasourceCatalogAssetDto `tfsdk:"data"`
	TotalElements *int64                       `tfsdk:"total_elements"`
}

type CatalogFilterDto struct {
	// Children *[]FilterElementDto `tfsdk:"children"`
	Id    *string `tfsdk:"id"`
	Name  *string `tfsdk:"name"`
	Query *string `tfsdk:"query"`
}

type DatasourceSearchDto struct {
	CatalogFilters []CatalogFilterDto                        `tfsdk:"catalog_filters"`
	SearchRules    SearchCollectionDatasourceCatalogAssetDto `tfsdk:"search_rules"`
}
