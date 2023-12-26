package datasource_datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TimeZoneDto struct {
	TimeZone  *string `tfsdk:"timezone"`
	UtcOffset *string `tfsdk:"utc_offset"`
}

type BigQueryParams struct {
	Type             types.String `tfsdk:"type"`
	BillingProjectID *string      `tfsdk:"billing_project_id"`
	DatasetID        *string      `tfsdk:"dataset_id"`
	ProjectID        *string      `tfsdk:"project_id"`
	TimezoneData     TimeZoneDto  `tfsdk:"timezone_data"`
}

type CreateDatasourceDto struct {
	ID             types.String    `tfsdk:"id"`
	Name           *string         `tfsdk:"name"`
	CronExpression *string         `tfsdk:"cron_expression"`
	Type           types.String    `tfsdk:"type"`
	SecretID       *string         `tfsdk:"secret_id"`
	BigQuery       *BigQueryParams `tfsdk:"bigquery"`
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
			"bigquery": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"billing_project_id": schema.StringAttribute{
						Required: true,
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
						Required: true,
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
