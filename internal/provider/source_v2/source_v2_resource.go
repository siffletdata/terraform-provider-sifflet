package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/datasource"
	"terraform-provider-sifflet/internal/provider/source/parameters"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource               = &sourceV2Resource{}
	_ resource.ResourceWithConfigure  = &sourceV2Resource{}
	_ resource.ResourceWithModifyPlan = &sourceV2Resource{}
	// _ resource.ResourceWithMoveState  = &sourceV2Resource{}
)

func (r sourceV2Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// If the request is planned for destruction, do nothing.
		return
	}

	var plan sourceV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: change to v2 parameters model
	var parametersModel parameters.ParametersModel
	resp.Diagnostics.Append(plan.Parameters.As(ctx, &parametersModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceType, err := parametersModel.GetSourceType()
	if err != nil {
		// not adding an error diagnostic here (the source type may still be unknown at that point, for instance if dynamic blocks are used).
		return
	}

	diags = resp.Plan.SetAttribute(ctx, path.Root("parameters").AtName("source_type"), sourceType.SchemaSourceType())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func newSourceV2Resource() resource.Resource {
	return &sourceV2Resource{}
}

type sourceV2Resource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *sourceV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_v2"
}

func sourceV2ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "A Sifflet source.",
		MarkdownDescription: `A Sifflet source. A source is any system that's monitored by Sifflet.

~> Consider adding a ` + "`lifecycle { prevent_destroy = true }` to `sifflet_source_v2`" + ` resources once they are correctly configured. Deleting a source deletes all associated data, including monitors on that source.
		`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the source.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"name": schema.StringAttribute{
				Description: "Source name.",
				Required:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone for the source. If empty, defaults to UTC.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UTC"),
			},
			// TODO: change to v2 parameters model
			"parameters": parameters.ParametersModel{}.TerraformSchema(),
		},
	}

}

func (r *sourceV2Resource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = sourceV2ResourceSchema(ctx)
}

func (r *sourceV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// No default timeout, this resource implements its own timeouts.

	var plan sourceV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, tfutils.DefaultTimeouts.Create)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	sourceDto, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dtoBytes, marshalErr := json.Marshal(sourceDto)
	if marshalErr != nil {
		resp.Diagnostics.AddError("Unable to serialize source", marshalErr.Error())
		return
	}

	// QUESTION: how to create the PublicCreateSourceV2JSONBody with the dtoBytes object ?
	var body sifflet.PublicCreateSourceV2JSONBody

	sourceResponse, err := r.client.PublicCreateSourceV2WithResponse(ctx, sifflet.PublicCreateSourceV2JSONRequestBody(body))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create source", err.Error())
		return
	}

	if sourceResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create source", sourceResponse.StatusCode(), sourceResponse.Body,
		)
		return
	}

	// QUESTION: how to get the json.RawMessage from sourceResponse.JSON201 and change it to PublicGetSourceV2Dto ?

	var newState sourceV2Model
	diags = newState.FromDto(ctx, *sourceResponse.JSON201)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState.Timeouts = plan.Timeouts

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No default timeout, this resource implements its own timeouts.

	var state sourceV2Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, tfutils.DefaultTimeouts.Read)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, diags := state.ModelId()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PublicGetSourceV2WithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source: could not parse API response", err.Error())
		return
	}

	if res.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read source", res.StatusCode(), res.Body)
		return
	}

	// QUESTION: how to get the json.RawMessage from sourceResponse.JSON201 and change it to PublicGetSourceV2Dto ?
	var newState sourceV2Model
	diags = newState.FromDto(ctx, *res.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState.Timeouts = state.Timeouts

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No default timeout, this resource implements its own timeouts.

	var plan sourceV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, tfutils.DefaultTimeouts.Update)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	id, diags := plan.ModelId()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := plan.ToUpdateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// QUESTION: how to create the PublicEditSourceV2JSONRequestBody object ?
	updateResponse, err := r.client.PublicEditSourceV2WithResponse(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update source", err.Error())
		return
	}

	if updateResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update source", updateResponse.StatusCode(), updateResponse.Body,
		)
		return
	}

	// QUESTION: how to get the json.RawMessage from sourceResponse.JSON201 and change it to PublicGetSourceV2Dto ?
	var newState sourceV2Model
	diags = newState.FromDto(ctx, *updateResponse.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState.Timeouts = plan.Timeouts

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No default timeout, this resource implements its own timeouts.

	var state sourceV2Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, tfutils.DefaultTimeouts.Delete)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	id, diags := state.ModelId()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialResponse, _ := r.client.PublicDeleteSourceV2WithResponse(ctx, id)

	if credentialResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete source",
			credentialResponse.StatusCode(), credentialResponse.Body,
		)
		return
	}

}

func (r *sourceV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *sourceV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*apiclients.HttpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *HttpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = clients.Client
}

func (r sourceV2Resource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	types := parameters.GetAllSourceTypes()
	paths := make([]path.Expression, len(types))
	for i, sourceType := range types {
		paths[i] = path.MatchRoot("parameters").AtName(sourceType)
	}

	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			paths...,
		),
	}
}

func moveTimezone(sourceTz *datasource.TimeZoneDto) types.String {
	if sourceTz == nil {
		return types.StringNull()
	}
	if sourceTz.TimeZone.ValueString() != "" {
		return sourceTz.TimeZone
	}
	return sourceTz.UtcOffset
}

// MoveState moves the state from the deprecated sifflet_datasource resource to the sifflet_source resource.
// Remove this method once the sifflet_datasource resource is removed.
// TODO: be able to move state from sifflet_source to sifflet_source_v2
// func (r *sourceV2Resource) MoveState(ctx context.Context) []resource.StateMover {
// 	sourceSchema := datasource.DatasourceResourceSchema(ctx)
// 	return []resource.StateMover{
// 		{
// 			SourceSchema: &sourceSchema,
// 			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
// 				if req.SourceTypeName != "sifflet_datasource" {
// 					return
// 				}

// 				if req.SourceSchemaVersion != 0 {
// 					return
// 				}

// 				var sourceStateData datasource.CreateDatasourceDto
// 				resp.Diagnostics.Append(req.SourceState.Get(ctx, &sourceStateData)...)
// 				if resp.Diagnostics.HasError() {
// 					return
// 				}

// 				var parametersModel parameters.ParametersModel
// 				var timezone types.String
// 				var diags diag.Diagnostics
// 				if sourceStateData.BigQuery != nil {
// 					parametersModel, diags = parameters.BigQueryParametersModel{
// 						ProjectId:        sourceStateData.BigQuery.ProjectID,
// 						BillingProjectId: sourceStateData.BigQuery.BillingProjectID,
// 					}.AsParametersModel(ctx)
// 					timezone = moveTimezone(sourceStateData.BigQuery.TimezoneData)
// 				} else if sourceStateData.DBT != nil {
// 					parametersModel, diags = parameters.DbtParametersModel{
// 						ProjectName: sourceStateData.DBT.ProjectName,
// 						Target:      sourceStateData.DBT.Target,
// 					}.AsParametersModel(ctx)
// 					timezone = moveTimezone(sourceStateData.DBT.TimezoneData)
// 				} else if sourceStateData.Snowflake != nil {
// 					parametersModel, diags = parameters.SnowflakeParametersModel{
// 						AccountIdentifier: sourceStateData.Snowflake.AccountIdentifier,
// 						Database:          sourceStateData.Snowflake.Database,
// 						Schema:            sourceStateData.Snowflake.Schema,
// 						Warehouse:         sourceStateData.Snowflake.Warehouse,
// 					}.AsParametersModel(ctx)
// 					timezone = moveTimezone(sourceStateData.Snowflake.TimezoneData)
// 				} else {
// 					resp.Diagnostics.AddError("Unsupported source type", "The sifflet_datasource type is not supported for this move operation.")
// 					return
// 				}
// 				resp.Diagnostics.Append(diags...)
// 				if diags.HasError() {
// 					return
// 				}

// 				parameters, diags := types.ObjectValueFrom(ctx, parametersModel.AttributeTypes(), parametersModel)
// 				resp.Diagnostics.Append(diags...)
// 				if diags.HasError() {
// 					return
// 				}

// 				t := types.ObjectNull(
// 					map[string]attr.Type{
// 						"create": types.StringType,
// 						"read":   types.StringType,
// 						"update": types.StringType,
// 						"delete": types.StringType,
// 					},
// 				)

// 				targetStateData := sourceV2Model{
// 					baseSourceV2Model: baseSourceV2Model{
// 						ID:         sourceStateData.ID,
// 						Name:       sourceStateData.Name,
// 						Timezone:   timezone,
// 						Parameters: parameters,
// 					},
// 					Timeouts: timeouts.Value{Object: t},
// 				}

// 				resp.Diagnostics.Append(resp.TargetState.Set(ctx, targetStateData)...)
// 			},
// 		},
// 	}
// }
