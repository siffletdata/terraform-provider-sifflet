package datasource_datasource

import (
	"context"

	data_source "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ErrorMessage struct {
	Title  string `json:"title"`
	Status int64  `json:"status"`
	Detail string `json:"detail"`
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
		Description:         "Read all Sifflet tags. **Deprecated: This data source relies on an unstable API that may change in the future.**",
		MarkdownDescription: "Read all Sifflet tags. **Deprecated: This data source relies on an unstable API that may change in the future.**",
		DeprecationMessage:  "This data source relies on an unstable API that may change in the future.",
		Attributes: map[string]data_source.Attribute{
			"data": data_source.ListNestedAttribute{
				Description: "List of all tags.",
				NestedObject: data_source.NestedAttributeObject{
					Attributes: map[string]data_source.Attribute{
						"created_by": data_source.StringAttribute{
							Description:        "Username that created the tag.",
							Computed:           true,
							DeprecationMessage: "This attribute may not be available in future versions of the Sifflet API. We recommend not relying on it.",
						},
						"created_date": data_source.StringAttribute{
							Description:        "Date of tag creation.",
							Computed:           true,
							DeprecationMessage: "This attribute may not be available in future versions of the Sifflet API. We recommend not relying on it.",
						},
						"description": data_source.StringAttribute{
							Description: "Tag description.",
							Computed:    true,
						},
						"editable": data_source.BoolAttribute{
							Computed:           true,
							Description:        "If tag is editable.",
							DeprecationMessage: "This attribute may not be available in future versions of the Sifflet API. We recommend not relying on it.",
						},
						"id": data_source.StringAttribute{
							Computed:    true,
							Description: "Tag UID",
						},
						"last_modified_date": data_source.StringAttribute{
							Description:        "Date of tag last modification.",
							Computed:           true,
							DeprecationMessage: "This attribute may not be available in future versions of the Sifflet API. We recommend not relying on it.",
						},
						"modified_by": data_source.StringAttribute{
							Computed:           true,
							Description:        "Last username that modified the tag.",
							DeprecationMessage: "This attribute may not be available in future versions of the Sifflet API. We recommend not relying on it.",
						},
						"name": data_source.StringAttribute{
							Computed:    true,
							Description: "Name to represent your integration in Sifflet.",
						},
						"type": data_source.StringAttribute{
							Computed:           true,
							Description:        "Type of tag (ie: GENERIC, HIDDEN_DATA_CLASSIFICATION, VISIBLE_DATA_CLASSIFICATION, TERM, BIGQUERY_EXTERNAL, SNOWFLAKE_EXTERNAL)",
							DeprecationMessage: "This attribute may not be available in future versions of the Sifflet API. We recommend not relying on it.",
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
