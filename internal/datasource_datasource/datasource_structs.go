package datasource_datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	data_source "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	BillingProjectID types.String `tfsdk:"billing_project_id"`
	DatasetID        types.String `tfsdk:"dataset_id"`
	ProjectID        types.String `tfsdk:"project_id"`
	TimezoneData     *TimeZoneDto `tfsdk:"timezone_data"`
}

type DBTParams struct {
	Type         types.String `tfsdk:"type"`
	ProjectName  types.String `tfsdk:"project_name"`
	Target       types.String `tfsdk:"target"`
	TimezoneData *TimeZoneDto `tfsdk:"timezone_data"`
}

type SnowflakeParams struct {
	Type              types.String `tfsdk:"type"`
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	Database          types.String `tfsdk:"database"`
	Schema            types.String `tfsdk:"schema"`
	Warehouse         types.String `tfsdk:"warehouse"`
	TimezoneData      *TimeZoneDto `tfsdk:"timezone_data"`
}

type CreateDatasourceDto struct {
	ID             types.String     `tfsdk:"id"`
	Name           types.String     `tfsdk:"name"`
	CronExpression *string          `tfsdk:"cron_expression"`
	Type           types.String     `tfsdk:"type"`
	SecretID       types.String     `tfsdk:"secret_id"`
	BigQuery       *BigQueryParams  `tfsdk:"bigquery"`
	DBT            *DBTParams       `tfsdk:"dbt"`
	Snowflake      *SnowflakeParams `tfsdk:"snowflake"`
	CreatedBy      types.String     `tfsdk:"created_by"`
	CreatedDate    types.String     `tfsdk:"created_date"`
	ModifiedBy     types.String     `tfsdk:"modified_by"`
	Tags           *[]types.String  `tfsdk:"tags"`
}

type ErrorMessage struct {
	Title  string `json:"title"`
	Status int64  `json:"status"`
	Detail string `json:"detail"`
}

func DatasourceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Create a Sifflet data source. **Deprecated: this resource relies on unstable APIs that may change in the future, and doesn't support all Sifflet source types. Use the `sifflet_source` resource instead**.",
		MarkdownDescription: "Create a Sifflet data source. **Deprecated: this resource relies on unstable APIs that may change in the future, and doesn't support all Sifflet source types. Use the `sifflet_source` resource instead**.",
		DeprecationMessage:  "This resource relies on unstable APIs that may change in the future, and doesn't support all Sifflet source types. Use the `sifflet_source` resource instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Data source UID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name to represent your integration in Sifflet",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cron_expression": schema.StringAttribute{
				Optional:    true,
				Description: "Cron expression used to defined schedule refresh of the data source.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of data source (ie: dbt, bigquery)",
			},
			"secret_id": schema.StringAttribute{
				Optional:    true,
				Description: "Secret ID used by the connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "Username that created the data source.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_date": schema.StringAttribute{
				Computed:    true,
				Description: "Date of data source creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "Last username that modified the datasource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of tags UUID.",
			},
			"bigquery": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:    true,
						Description: "Type of data source (ie: dbt, bigquery).",
					},
					"billing_project_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "GCP project used for billing.",
						Default:     stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"dataset_id": schema.StringAttribute{
						Required:    true,
						Description: "BigQuery dataset to add as data source.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"project_id": schema.StringAttribute{
						Required:    true,
						Description: "Project where the dataset to add is located.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"timezone_data": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timezone informations of your data source.",
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
								Required:    true,
								Description: "Timezone of your data source (ie: UTC).",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"utc_offset": schema.StringAttribute{
								Required:    true,
								Description: "Timezone offset of your data source (ie: '(UTC+00:00)').",
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
						Computed:    true,
						Description: "Type of data source (ie: dbt, bigquery).",
					},
					"project_name": schema.StringAttribute{
						Required:    true,
						Description: "The name of your dbt project (the 'name' in your dbt_project.yml file).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"target": schema.StringAttribute{
						Required:    true,
						Description: "the target value of the profile that corresponds to your project (the 'target' in your profiles.yml).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"timezone_data": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timezone informations of your data source.",
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
								Required:    true,
								Description: "Timezone of your data source (ie: UTC).",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"utc_offset": schema.StringAttribute{
								Required:    true,
								Description: "Timezone offset of your data source (ie: '(UTC+00:00)').",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"snowflake": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:    true,
						Description: "Type of data source (ie: dbt, bigquery).",
					},
					"account_identifier": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Snowflake account identifier (see: https://docs.siffletdata.com/docs/snowflake#3--create-the-snowflake-connection-using-sifflets-integrations-page).",
						Default:     stringdefault.StaticString(""),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"database": schema.StringAttribute{
						Required:    true,
						Description: "Snowflake database.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"schema": schema.StringAttribute{
						Required:    true,
						Description: "Snowflake schema.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"warehouse": schema.StringAttribute{
						Required:    true,
						Description: "Snowflake warehouse.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"timezone_data": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timezone informations of your data source.",
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
								Required:    true,
								Description: "Timezone of your data source (ie: UTC).",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"utc_offset": schema.StringAttribute{
								Required:    true,
								Description: "Timezone offset of your data source (ie: '(UTC+00:00)').",
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
	Snowflake        *SnowflakeParams                    `tfsdk:"snowflake"`
	// Tags             *[]TagDto                           `tfsdk:"tags"`
	Type string `tfsdk:"type"`
}

type SearchCollectionDatasourceCatalogAssetDto struct {
	Data          *[]DatasourceCatalogAssetDto `tfsdk:"data"`
	TotalElements *int64                       `tfsdk:"total_elements"`
}

type DatasourceSearchDto struct {
	SearchRules SearchCollectionDatasourceCatalogAssetDto `tfsdk:"search_rules"`
}

func DatasourceDataSourceSchema(ctx context.Context) data_source.Schema {
	return data_source.Schema{
		Description:         "Read all Sifflet Data Sources. **Deprecated: this data source doesn't support all Sifflet source types. Use the `sifflet_sources` data source instead.**",
		MarkdownDescription: "Read all Sifflet Data Sources. **Deprecated: this data source doesn't support all Sifflet source types. Use the `sifflet_sources` data source instead.**",
		DeprecationMessage:  "This data source relies on an unstable API that may change in the future. It also doesn't support all Sifflet source types. Use the `sifflet_sources` data source instead.",
		Attributes: map[string]data_source.Attribute{
			"search_rules": data_source.SingleNestedAttribute{
				Attributes: map[string]data_source.Attribute{
					"data": data_source.ListNestedAttribute{
						Description: "List of all data sources.",
						NestedObject: data_source.NestedAttributeObject{
							Attributes: map[string]data_source.Attribute{
								"created_by": data_source.StringAttribute{
									Description: "Username that created the data source.",
									Computed:    true,
								},
								"created_date": data_source.Int64Attribute{
									Description: "Date of data source creation.",
									Computed:    true,
								},
								"cron_expression": data_source.StringAttribute{
									Description: "Cron expression used to defined schedule refresh of the data source.",
									Computed:    true,
								},
								"entity_type": data_source.StringAttribute{
									Computed:    true,
									Description: "Sifflet entity type (ie: DATASOURCE).",
								},
								"id": data_source.StringAttribute{
									Computed:    true,
									Description: "Data source UID",
								},
								"last_modified_date": data_source.Int64Attribute{
									Description: "Date of data source last modification.",
									Computed:    true,
								},
								"modified_by": data_source.StringAttribute{
									Computed:    true,
									Description: "Last username that modified the datasource.",
								},
								"name": data_source.StringAttribute{
									Computed:    true,
									Description: "Name to represent your integration in Sifflet.",
								},
								"next_execution": data_source.Int64Attribute{
									Computed:    true,
									Description: "Date of the next refresh for the data source.",
								},
								"type": data_source.StringAttribute{
									Computed:    true,
									Description: "Type of data source (ie: dbt, bigquery)",
								},
								"bigquery": data_source.SingleNestedAttribute{
									Computed: true,
									Optional: true,
									Attributes: map[string]data_source.Attribute{
										"type": data_source.StringAttribute{
											Computed:    true,
											Description: "Type of data source (ie: dbt, bigquery).",
										},
										"billing_project_id": data_source.StringAttribute{
											Computed:    true,
											Description: "GCP project used for billing.",
										},
										"dataset_id": data_source.StringAttribute{
											Computed:    true,
											Description: "BigQuery dataset to add as data source.",
										},
										"project_id": data_source.StringAttribute{
											Computed:    true,
											Description: "Project where the dataset to add is located.",
										},
										"timezone_data": data_source.SingleNestedAttribute{
											Optional:    true,
											Computed:    true,
											Description: "Timezone informations of your data source.",
											Attributes: map[string]data_source.Attribute{
												"timezone": data_source.StringAttribute{
													Computed:    true,
													Description: "Timezone of your data source (ie: UTC).",
												},
												"utc_offset": data_source.StringAttribute{
													Computed:    true,
													Description: "Timezone offset of your data source (ie: '(UTC+00:00)').",
												},
											},
										},
									},
								},
								"dbt": data_source.SingleNestedAttribute{
									Optional: true,
									Computed: true,
									Attributes: map[string]data_source.Attribute{
										"type": data_source.StringAttribute{
											Computed:    true,
											Description: "Type of data source (ie: dbt, bigquery).",
										},
										"project_name": data_source.StringAttribute{
											Computed:    true,
											Description: "The name of your dbt project (the 'name' in your dbt_project.yml file).",
										},
										"target": data_source.StringAttribute{
											Computed:    true,
											Description: "the target value of the profile that corresponds to your project (the 'target' in your profiles.yml).",
										},
										"timezone_data": data_source.SingleNestedAttribute{
											Optional:    true,
											Computed:    true,
											Description: "Timezone informations of your data source.",
											Attributes: map[string]data_source.Attribute{
												"timezone": data_source.StringAttribute{
													Computed:    true,
													Description: "Timezone of your data source (ie: UTC).",
												},
												"utc_offset": data_source.StringAttribute{
													Computed:    true,
													Description: "Timezone offset of your data source (ie: '(UTC+00:00)').",
												},
											},
										},
									},
								},
								"snowflake": data_source.SingleNestedAttribute{
									Optional: true,
									Computed: true,
									Attributes: map[string]data_source.Attribute{
										"type": data_source.StringAttribute{
											Computed:    true,
											Description: "Type of data source (ie: dbt, bigquery).",
										},
										"account_identifier": data_source.StringAttribute{
											Computed:    true,
											Description: "Snowflake account identifier (see: https://docs.siffletdata.com/docs/snowflake#3--create-the-snowflake-connection-using-sifflets-integrations-page).",
										},
										"database": data_source.StringAttribute{
											Computed:    true,
											Description: "Snowflake database.",
										},
										"schema": data_source.StringAttribute{
											Computed:    true,
											Description: "Snowflake schema.",
										},
										"warehouse": data_source.StringAttribute{
											Computed:    true,
											Description: "Snowflake warehouse.",
										},
										"timezone_data": data_source.SingleNestedAttribute{
											Optional:    true,
											Computed:    true,
											Description: "Timezone informations of your data source.",
											Attributes: map[string]data_source.Attribute{
												"timezone": data_source.StringAttribute{
													Computed:    true,
													Description: "Timezone of your data source (ie: UTC).",
												},
												"utc_offset": data_source.StringAttribute{
													Computed:    true,
													Description: "Timezone offset of your data source (ie: '(UTC+00:00)').",
												},
											},
										},
									},
								},
							},
						},
						Computed: true,
					},
					"total_elements": data_source.Int64Attribute{
						Computed:    true,
						Description: "Total number of data sources.",
					},
				},
				Computed: true,
			},
		},
	}
}
