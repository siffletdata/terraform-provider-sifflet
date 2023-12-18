package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	Type             *string     `tfsdk:"type"`
	BillingProjectID *string     `tfsdk:"billing_project_id"`
	DatasetID        *string     `tfsdk:"dataset_id"`
	ProjectID        *string     `tfsdk:"project_id"`
	TimezoneData     TimeZoneDto `tfsdk:"timezone_data"`
}

type CreateDatasourceDto struct {
	ID       types.String    `tfsdk:"id"`
	Name     *string         `tfsdk:"name"`
	Type     *string         `tfsdk:"type"`
	SecretID *string         `tfsdk:"secret_id"`
	BigQuery *BigQueryParams `tfsdk:"bigquery"`
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
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"secret_id": schema.StringAttribute{
				Optional: true,
			},
			"bigquery": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
					},
					"billing_project_id": schema.StringAttribute{
						Required: true,
					},
					"dataset_id": schema.StringAttribute{
						Required: true,
					},
					"project_id": schema.StringAttribute{
						Required: true,
					},
					"timezone_data": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"timezone": schema.StringAttribute{
								Required: true,
							},
							"utc_offset": schema.StringAttribute{
								Required: true,
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
	jsonData := []byte(fmt.Sprintf(`{"type": "%s"}`, *plan.BigQuery.Type))
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
	}

	// Create new order
	datasourceResponse, err := r.client.CreateDatasource(ctx, datasource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	resBody, _ := io.ReadAll(datasourceResponse.Body)
	tflog.Debug(ctx, "test1 "+string(resBody))

	// Map response body to schema and populate Computed attribute values
	// plan.ID = types.StringValue(strconv.Itoa(order.ID))
	// for orderItemIndex, orderItem := range order.Items {
	// 	plan.Items[orderItemIndex] = orderItemModel{
	// 		Coffee: orderItemCoffeeModel{
	// 			ID:          types.Int64Value(int64(orderItem.Coffee.ID)),
	// 			Name:        types.StringValue(orderItem.Coffee.Name),
	// 			Teaser:      types.StringValue(orderItem.Coffee.Teaser),
	// 			Description: types.StringValue(orderItem.Coffee.Description),
	// 			Price:       types.Float64Value(orderItem.Coffee.Price),
	// 			Image:       types.StringValue(orderItem.Coffee.Image),
	// 		},
	// 		Quantity: types.Int64Value(int64(orderItem.Quantity)),
	// 	}
	// }
	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *datasourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *datasourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
