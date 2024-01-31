package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	sifflet "terraform-provider-sifflet/internal/client"
	datasource_struct "terraform-provider-sifflet/internal/datasource_datasource"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &datasourceResource{}
	_ resource.ResourceWithConfigure = &datasourceResource{}
)

// NewDataSourceResource is a helper function to simplify the provider implementation.
func NewDataSourceResource() resource.Resource {
	return &datasourceResource{}
}

// datasourceResource is the resource implementation.
type datasourceResource struct {
	client *sifflet.Client
}

// Metadata returns the resource type name.
func (r *datasourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasource"
}

// Schema defines the schema for the resource.
func (r *datasourceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = datasource_struct.DatasourceResourceSchema(ctx)
}

// Create creates the resource and sets the initial Terraform state.
func (r *datasourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// TODO: Datasources is not tested, can be create with anythings as value

	var plan datasource_struct.CreateDatasourceDto
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := sifflet.CreateDatasourceDto_Params{}

	var jsonData []byte
	var connect_type string

	// Assuming you have some JSON data, you can unmarshal it into the RawMessage field
	if plan.BigQuery != nil {
		connect_type = "bigquery"

		jsonData = []byte(fmt.Sprintf(`
	{
		"type": "%s",
		"billingProjectId": "%s",
		"datasetId": "%s",
		"projectId": "%s",
		"timezoneData": {
			"timezone": %s,
			"utcOffset": %s
		}
	}
	`,
			connect_type,
			plan.BigQuery.BillingProjectID,
			plan.BigQuery.DatasetID,
			plan.BigQuery.ProjectID,
			plan.BigQuery.TimezoneData.TimeZone,
			plan.BigQuery.TimezoneData.UtcOffset,
		))
		tflog.Debug(ctx, "Params:  "+string(jsonData))
	} else if plan.DBT != nil {
		connect_type = "dbt"

		jsonData = []byte(fmt.Sprintf(`
	{
		"type": "%s",
		"projectName": "%s",
		"target": "%s",
		"timezoneData": {
			"timezone": %s,
			"utcOffset": %s
		}
	}
	`,
			connect_type,
			plan.DBT.ProjectName,
			plan.DBT.Target,
			plan.DBT.TimezoneData.TimeZone,
			plan.DBT.TimezoneData.UtcOffset,
		))
		tflog.Debug(ctx, "Params:  "+string(jsonData))
	} else if plan.Snowflake != nil {
		connect_type = "snowflake"

		jsonData = []byte(fmt.Sprintf(`
	{
		"type": "%s",
		"accountIdentifier": "%s",
		"database": "%s",
		"schema": "%s",
		"warehouse": "%s",
		"timezoneData": {
			"timezone": %s,
			"utcOffset": %s
		}
	}
	`,
			connect_type,
			plan.Snowflake.AccountIdentifier,
			plan.Snowflake.Database,
			plan.Snowflake.Schema,
			plan.Snowflake.Warehouse,
			plan.Snowflake.TimezoneData.TimeZone,
			plan.Snowflake.TimezoneData.UtcOffset,
		))
		tflog.Debug(ctx, "Params:  "+string(jsonData))
	}

	err := json.Unmarshal(jsonData, &params)
	if err != nil {
		fmt.Println("Error unmarshaling JSON from plan:", err)
		return
	}

	var tags []openapi_types.UUID

	for _, tag := range *plan.Tags {
		tags = append(tags, uuid.MustParse(tag.String()))
	}

	// Generate API request body from plan
	datasource := sifflet.CreateDatasourceJSONRequestBody{
		Params:         params,
		CronExpression: plan.CronExpression,
		Type:           connect_type,
		Tags:           &tags,
	}
	datasource.Name = plan.Name.String()
	datasource.SecretId = plan.SecretID.ValueStringPointer()

	// Create new order
	datasourceResponse, _ := r.client.CreateDatasource(ctx, datasource)

	resBody, _ := io.ReadAll(datasourceResponse.Body)
	tflog.Debug(ctx, "Response:  "+string(resBody))

	if datasourceResponse.StatusCode != http.StatusCreated {
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

	var result sifflet.DatasourceDto
	if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
		resp.Diagnostics.AddError(
			"Can not unmarshal JSON",
			err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(result.Id.String())
	plan.Name = types.StringValue(result.Name)
	plan.CronExpression = result.CronExpression
	plan.Type = types.StringValue(result.Type)
	plan.CreatedBy = types.StringValue(*result.CreatedBy)
	plan.CreatedDate = types.StringValue(strconv.FormatInt(*result.CreatedDate, 10))
	plan.ModifiedBy = types.StringValue(*result.ModifiedBy)
	plan.SecretID = types.StringValue(*result.SecretId)

	if plan.BigQuery != nil {
		resultParams, _ := result.Params.AsBigQueryParams()
		plan.BigQuery.BillingProjectID = types.StringValue(*resultParams.BillingProjectId)
		plan.BigQuery.ProjectID = types.StringValue(*resultParams.ProjectId)
		plan.BigQuery.DatasetID = types.StringValue(*resultParams.DatasetId)
		plan.BigQuery.Type = types.StringValue(resultParams.Type)
		plan.BigQuery.TimezoneData.TimeZone = types.StringValue(resultParams.TimezoneData.Timezone)
		plan.BigQuery.TimezoneData.UtcOffset = types.StringValue(resultParams.TimezoneData.UtcOffset)
	} else if plan.DBT != nil {
		resultParams, _ := result.Params.AsDBTParams()
		plan.DBT.Target = types.StringValue(*resultParams.Target)
		plan.DBT.ProjectName = types.StringValue(*resultParams.ProjectName)
		plan.DBT.Type = types.StringValue(resultParams.Type)
		plan.DBT.TimezoneData.TimeZone = types.StringValue(resultParams.TimezoneData.Timezone)
		plan.DBT.TimezoneData.UtcOffset = types.StringValue(resultParams.TimezoneData.UtcOffset)

	} else if plan.Snowflake != nil {
		resultParams, _ := result.Params.AsSnowflakeParams()
		plan.Snowflake.AccountIdentifier = types.StringValue(*resultParams.AccountIdentifier)
		plan.Snowflake.Database = types.StringValue(*resultParams.Database)
		plan.Snowflake.Schema = types.StringValue(*resultParams.Schema)
		plan.Snowflake.Warehouse = types.StringValue(*resultParams.Warehouse)
		plan.Snowflake.Type = types.StringValue(resultParams.Type)
		plan.Snowflake.TimezoneData.TimeZone = types.StringValue(resultParams.TimezoneData.Timezone)
		plan.Snowflake.TimezoneData.UtcOffset = types.StringValue(resultParams.TimezoneData.UtcOffset)

	}

	var result_tags []basetypes.StringValue
	for _, tag := range *result.Tags {
		Id := tag.Id.String()
		result_tags = append(result_tags, types.StringValue(Id))
	}

	plan.Tags = &result_tags
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *datasourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datasource_struct.CreateDatasourceDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.String()

	itemResponse, err := r.client.GetDatasourceById(ctx, uuid.MustParse(id))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Item",
			err.Error(),
		)
		return
	}

	resBody, _ := io.ReadAll(itemResponse.Body)
	tflog.Debug(ctx, fmt.Sprintf("Response: %d ", itemResponse.Body))

	if itemResponse.StatusCode == http.StatusNotFound {
		// TODO: in case of 404 nothing is return by the API
		resp.State.RemoveResource(ctx)
		return
	}

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

	var result sifflet.DatasourceDto
	if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
		resp.Diagnostics.AddError(
			"Can not unmarshal JSON",
			err.Error(),
		)
		return
	}

	state = datasource_struct.CreateDatasourceDto{
		ID:             types.StringValue(result.Id.String()),
		Name:           types.StringValue(result.Name),
		CreatedBy:      types.StringValue(*result.CreatedBy),
		CreatedDate:    types.StringValue(strconv.FormatInt(*result.CreatedDate, 10)),
		ModifiedBy:     types.StringValue(*result.ModifiedBy),
		CronExpression: result.CronExpression,
		Type:           types.StringValue(result.Type),
		SecretID:       types.StringValue(*result.SecretId),
	}

	if result.Type == "bigquery" {
		resultParams, _ := result.Params.AsBigQueryParams()

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

		state.BigQuery = &result_bq
	} else if result.Type == "dbt" {
		resultParams, _ := result.Params.AsDBTParams()

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

		state.DBT = &result_dbt
	} else if result.Type == "snowflake" {
		resultParams, _ := result.Params.AsSnowflakeParams()

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

		state.Snowflake = &result_snowflake
	}

	var result_tags []basetypes.StringValue
	for _, tag := range *result.Tags {
		Id := tag.Id.String()
		result_tags = append(result_tags, types.StringValue(Id))
	}
	state.Tags = &result_tags

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *datasourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// NOT IMPLEMENTED IN OPENAPI CONTRACT
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *datasourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state datasource_struct.CreateDatasourceDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.String()

	datasourceResponse, _ := r.client.DeleteDatasourceById(ctx, uuid.MustParse(id))
	resBody, _ := io.ReadAll(datasourceResponse.Body)
	tflog.Debug(ctx, "Response "+string(resBody))

	if datasourceResponse.StatusCode != http.StatusNoContent {
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

}

func (r *datasourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *datasourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *datasourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {

	var data datasource_struct.CreateDatasourceDto

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: maybe find something more elegant than chaining checks
	if data.DBT != nil && data.BigQuery != nil && data.Snowflake != nil {
		resp.Diagnostics.AddError(
			"Error",
			"Define only one type of data source",
		)
		return
	}
}
