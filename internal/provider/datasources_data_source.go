package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	sifflet "terraform-provider-sifflet/internal/alphaclient"
	datasource_struct "terraform-provider-sifflet/internal/datasource_datasource"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &datasourcesDataSource{}
	_ datasource.DataSourceWithConfigure = &datasourcesDataSource{}
)

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

	clients, ok := req.ProviderData.(*httpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *httpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = clients.AlphaClient
}

func (d *datasourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasources"
}

func (d *datasourcesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_struct.DatasourceDataSourceSchema(ctx)
}

func (d *datasourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datasource_struct.DatasourceSearchDto

	ItemsPerPage := int32(-1)

	params := sifflet.GetAllDatasourceParams{
		ItemsPerPage: &ItemsPerPage,
	}

	itemResponse, err := d.client.GetAllDatasource(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read data sources",
			err.Error(),
		)
		return
	}
	resBody, _ := io.ReadAll(itemResponse.Body)
	tflog.Debug(ctx, "Response:  "+string(resBody))

	if itemResponse.StatusCode != http.StatusOK {

		var message datasource_struct.ErrorMessage
		if err := json.Unmarshal(resBody, &message); err != nil { // Parse []byte to go struct pointer
			resp.Diagnostics.AddError(
				"Can not unmarshal JSON",
				err.Error(),
			)
			return
		}
		resp.Diagnostics.AddError(
			message.Title,
			message.Detail,
		)
		resp.State.RemoveResource(ctx)
		return
	}

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
		state.SearchRules.Data = &[]datasource_struct.DatasourceCatalogAssetDto{}
	}

	for _, data := range *result.SearchDatasources.Data {

		idString := data.Id.String()
		yEntityType := datasource_struct.DatasourceCatalogAssetDtoEntityType(data.EntityType)
		data_source_catalog_asset := datasource_struct.DatasourceCatalogAssetDto{
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
		}

		if data.Type == "bigquery" {

			resultParams, _ := data.Params.AsBigQueryParams()

			result_timezone := datasource_struct.TimeZoneDto{
				TimeZone:  types.StringValue(resultParams.TimezoneData.Timezone),
				UtcOffset: types.StringValue(resultParams.TimezoneData.UtcOffset),
			}

			result_bq := datasource_struct.BigQueryParams{
				Type:             types.StringValue(resultParams.Type),
				BillingProjectID: types.StringValue(*resultParams.BillingProjectId),
				DatasetID:        types.StringValue(*resultParams.DatasetId),
				ProjectID:        types.StringValue(*resultParams.ProjectId),
				TimezoneData:     &result_timezone,
			}
			data_source_catalog_asset.BigQuery = &result_bq
		} else if data.Type == "dbt" {
			resultParams, _ := data.Params.AsDBTParams()

			result_timezone := datasource_struct.TimeZoneDto{
				TimeZone:  types.StringValue(resultParams.TimezoneData.Timezone),
				UtcOffset: types.StringValue(resultParams.TimezoneData.UtcOffset),
			}

			result_dbt := datasource_struct.DBTParams{
				Type:         types.StringValue(resultParams.Type),
				Target:       types.StringValue(*resultParams.Target),
				ProjectName:  types.StringValue(*resultParams.ProjectName),
				TimezoneData: &result_timezone,
			}

			data_source_catalog_asset.DBT = &result_dbt
		} else if data.Type == "snowflake" {
			resultParams, _ := data.Params.AsSnowflakeParams()

			result_timezone := datasource_struct.TimeZoneDto{
				TimeZone:  types.StringValue(resultParams.TimezoneData.Timezone),
				UtcOffset: types.StringValue(resultParams.TimezoneData.UtcOffset),
			}

			result_snowflake := datasource_struct.SnowflakeParams{
				Type:              types.StringValue(resultParams.Type),
				Database:          types.StringValue(*resultParams.Database),
				Schema:            types.StringValue(*resultParams.Schema),
				Warehouse:         types.StringValue(*resultParams.Warehouse),
				AccountIdentifier: types.StringValue(*resultParams.AccountIdentifier),
				TimezoneData:      &result_timezone,
			}

			data_source_catalog_asset.Snowflake = &result_snowflake
		}

		*state.SearchRules.Data = append(*state.SearchRules.Data, data_source_catalog_asset)

	}

	for _, filters := range result.CatalogFilters {
		state.CatalogFilters = append(state.CatalogFilters, datasource_struct.CatalogFilterDto{
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
