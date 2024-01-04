package datasource_datasource

import (
	"context"

	data_source "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ErrorMessage struct {
	Title  string `json:"title"`
	Status int64  `json:"status"`
	Detail string `json:"detail"`
}

func TagResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Create a Sifflet tag.",
		Attributes: map[string]schema.Attribute{
			"created_by": schema.StringAttribute{
				Description: "Username that created the tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_date": schema.StringAttribute{
				Description: "Date of tag creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Tag description.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"editable": schema.BoolAttribute{
				Computed:    true,
				Description: "If tag is editable.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Tag UID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_date": schema.StringAttribute{
				Description: "Date of tag last modification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "Last username that modified the tag.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name to represent your integration in Sifflet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of tag (ie: GENERIC, HIDDEN_DATA_CLASSIFICATION, VISIBLE_DATA_CLASSIFICATION, TERM, BIGQUERY_EXTERNAL, SNOWFLAKE_EXTERNAL)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

type TagDtoType string

type TagDto struct {
	CreatedBy        types.String `tfsdk:"created_by"`
	CreatedDate      types.String `tfsdk:"created_date"`
	Description      *string      `tfsdk:"description"`
	Editable         types.Bool   `tfsdk:"editable"`
	Id               types.String `tfsdk:"id"`
	LastModifiedDate types.String `tfsdk:"last_modified_date"`
	ModifiedBy       types.String `tfsdk:"modified_by"`
	Name             *string      `tfsdk:"name"`
	Type             *TagDtoType  `tfsdk:"type"`
}

type SearchCollectionTagDto struct {
	Data          *[]TagDto `tfsdk:"data"`
	TotalElements *int64    `tfsdk:"total_elements"`
}

func TagDataSourceSchema(ctx context.Context) data_source.Schema {
	return data_source.Schema{
		Description: "Real all Sifflet tags.",
		Attributes: map[string]data_source.Attribute{
			"data": data_source.ListNestedAttribute{
				Description: "List of all tags.",
				NestedObject: data_source.NestedAttributeObject{
					Attributes: map[string]data_source.Attribute{
						"created_by": data_source.StringAttribute{
							Description: "Username that created the tag.",
							Computed:    true,
						},
						"created_date": data_source.StringAttribute{
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
						"last_modified_date": data_source.StringAttribute{
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
							Description: "Type of tag (ie: GENERIC, HIDDEN_DATA_CLASSIFICATION, VISIBLE_DATA_CLASSIFICATION, TERM, BIGQUERY_EXTERNAL, SNOWFLAKE_EXTERNAL)",
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
