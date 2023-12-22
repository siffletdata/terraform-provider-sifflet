package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &datasourceResource{}
	_ resource.ResourceWithConfigure = &datasourceResource{}
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
	ID       types.String    `tfsdk:"id"`
	Name     *string         `tfsdk:"name"`
	Type     types.String    `tfsdk:"type"`
	SecretID *string         `tfsdk:"secret_id"`
	BigQuery *BigQueryParams `tfsdk:"bigquery"`
}

type ErrorMessage struct {
	Title  string `json:"title"`
	Status int64  `json:"status"`
	Detail string `json:"detail"`
}

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
func (r *datasourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

// Create creates the resource and sets the initial Terraform state.
func (r *datasourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CreateDatasourceDto
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := sifflet.CreateDatasourceDto_Params{}

	// Assuming you have some JSON data, you can unmarshal it into the RawMessage field
	jsonData := []byte(fmt.Sprintf(`
	{
		"type": "bigquery",
		"billingProjectId": "%s",
		"datasetId": "%s",
		"projectId": "%s"
	}
	`, *plan.BigQuery.BillingProjectID,
		*plan.BigQuery.DatasetID,
		*plan.BigQuery.ProjectID,
	))
	tflog.Debug(ctx, "test1 "+string(jsonData))

	err := json.Unmarshal(jsonData, &params)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Generate API request body from plan
	datasource := sifflet.CreateDatasourceJSONRequestBody{
		Name:     *plan.Name,
		SecretId: plan.SecretID,
		Params:   params,
		Type:     "bigquery",
	}

	// Create new order
	datasourceResponse, err := r.client.CreateDatasource(ctx, datasource)

	resBody, _ := io.ReadAll(datasourceResponse.Body)
	tflog.Debug(ctx, "test1 "+string(resBody))

	if datasourceResponse.StatusCode != http.StatusCreated {
		var message ErrorMessage
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
	plan.Name = &result.Name
	plan.Type = types.StringValue(result.Type)
	plan.SecretID = result.SecretId
	resultParams, err := result.Params.AsBigQueryParams()
	plan.BigQuery.BillingProjectID = resultParams.BillingProjectId
	plan.BigQuery.ProjectID = resultParams.ProjectId
	plan.BigQuery.DatasetID = resultParams.DatasetId
	plan.BigQuery.Type = types.StringValue(resultParams.Type)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *datasourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CreateDatasourceDto
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
	tflog.Debug(ctx, "test1 "+string(resBody))

	var result sifflet.DatasourceDto
	if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
		resp.Diagnostics.AddError(
			"Can not unmarshal JSON",
			err.Error(),
		)
		return
	}

	resultParams, err := result.Params.AsBigQueryParams()

	result_timezone := TimeZoneDto{
		TimeZone:  &resultParams.TimezoneData.Timezone,
		UtcOffset: &resultParams.TimezoneData.UtcOffset,
	}

	result_bq := BigQueryParams{
		Type:             types.StringValue(resultParams.Type),
		BillingProjectID: resultParams.BillingProjectId,
		DatasetID:        resultParams.DatasetId,
		ProjectID:        resultParams.ProjectId,
		TimezoneData:     result_timezone,
	}

	state = CreateDatasourceDto{
		ID:       types.StringValue(result.Id.String()),
		Name:     &result.Name,
		Type:     types.StringValue(result.Type),
		SecretID: result.SecretId,
		BigQuery: &result_bq,
	}

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