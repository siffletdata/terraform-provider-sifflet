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

type SnowflakeParams struct {
	Type              types.String `tfsdk:"type"`
	AccountIdentifier *string      `tfsdk:"account_identifier"`
	Database          *string      `tfsdk:"database"`
	Schema            *string      `tfsdk:"schema"`
	Warehouse         *string      `tfsdk:"warehouse"`
	TimezoneData      *TimeZoneDto `tfsdk:"timezone_data"`
}

type CreateDatasourceDto struct {
	ID             types.String     `tfsdk:"id"`
	Name           *string          `tfsdk:"name"`
	CronExpression *string          `tfsdk:"cron_expression"`
	Type           types.String     `tfsdk:"type"`
	SecretID       *string          `tfsdk:"secret_id"`
	BigQuery       *BigQueryParams  `tfsdk:"bigquery"`
	DBT            *DBTParams       `tfsdk:"dbt"`
	Snowflake      *SnowflakeParams `tfsdk:"snowflake"`
	CreatedBy      types.String     `tfsdk:"created_by"`
	CreatedDate    types.String     `tfsdk:"created_date"`
	ModifiedBy     types.String     `tfsdk:"modified_by"`
}

type ErrorMessage struct {
	Title  string `json:"title"`
	Status int64  `json:"status"`
	Detail string `json:"detail"`
}

func DatasourceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
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

type TagDtoType string

type TagDto struct {
	CreatedBy        *string    `tfsdk:"created_by"`
	CreatedDate      *int64     `tfsdk:"created_date"`
	Description      *string    `tfsdk:"description"`
	Editable         *bool      `tfsdk:"editable"`
	Id               string     `tfsdk:"id"`
	LastModifiedDate *int64     `tfsdk:"last_modified_date"`
	ModifiedBy       *string    `tfsdk:"modified_by"`
	Name             string     `tfsdk:"name"`
	Type             TagDtoType `tfsdk:"type"`
}

type SearchCollectionTagDto struct {
	Data          *[]TagDto `tfsdk:"data"`
	TotalElements *int64    `tfsdk:"total_elements"`
}

func TagDataSourceSchema(ctx context.Context) data_source.Schema {
	return data_source.Schema{
		Attributes: map[string]data_source.Attribute{
			"data": data_source.ListNestedAttribute{
				Description: "List of all tags.",
				NestedObject: data_source.NestedAttributeObject{
					Attributes: map[string]data_source.Attribute{
						"created_by": data_source.StringAttribute{
							Description: "Username that created the tag.",
							Computed:    true,
						},
						"created_date": data_source.Int64Attribute{
							Description: "Date of tag creation.",
							Computed:    true,
						},
						"description": data_source.StringAttribute{
							Description: "Tag description.",
							Computed:    true,
						},
						"editable": data_source.BoolAttribute{
							Computed:    true,
							Description: "If tag is editable.",
						},
						"id": data_source.StringAttribute{
							Computed:    true,
							Description: "Tag UID",
						},
						"last_modified_date": data_source.Int64Attribute{
							Description: "Date of tag last modification.",
							Computed:    true,
						},
						"modified_by": data_source.StringAttribute{
							Computed:    true,
							Description: "Last username that modified the tag.",
						},
						"name": data_source.StringAttribute{
							Computed:    true,
							Description: "Name to represent your integration in Sifflet.",
						},
						"type": data_source.StringAttribute{
							Computed:    true,
							Description: "Type of tag (ie: GENERIC)",
						},
					},
				},
				Computed: true,
			},
			"total_elements": data_source.Int64Attribute{
				Computed:    true,
				Description: "Total number of tag.",
			},
		},
	}
}
