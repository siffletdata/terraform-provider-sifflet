package source

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/datasource"
	"terraform-provider-sifflet/internal/provider/source/parameters"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource               = &sourceResource{}
	_ resource.ResourceWithConfigure  = &sourceResource{}
	_ resource.ResourceWithModifyPlan = &sourceResource{}
	_ resource.ResourceWithMoveState  = &sourceResource{}
)

func (r sourceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// If the request is planned for destruction, do nothing.
		return
	}

	var plan sourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

func newSourceResource() resource.Resource {
	return &sourceResource{}
}

type sourceResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *sourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func SourceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "A Sifflet source.",
		MarkdownDescription: `A Sifflet source. A source is any system that's monitored by Sifflet.

~> Consider adding a ` + "`lifecycle { prevent_destroy = true }` to `sifflet_source`" + ` resources once they are correctly configured. Deleting a source deletes all associated data, including monitors on that source.
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
			"credentials": schema.StringAttribute{
				Description: "Name of the credentials used to connect to the source. Required for most datasources, except for 'athena', 'dbt' and 'quicksight' sources.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Source description.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Source name.",
				Required:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "Schedule for the source. Must be a valid cron expression. If empty, the source will only be refreshed when manually triggered.",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone for the source. If empty, defaults to UTC.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UTC"),
			},
			"tags": schema.ListNestedAttribute{
				Description: "Tags for the source. For each tag, you can provide either: an ID, a name, or a name + a kind (when the name alone is ambiguous). It's recommended to use tag IDs (coming from a `sifflet_tag` resource or data source) most of the time, but directly providing tag names can simplify some configurations.",
				Optional:    true,
				Computed:    true,
				Default: listdefault.StaticValue(types.ListValueMust(
					types.ObjectType{
						AttrTypes: tagModel{}.AttributeTypes(),
					},
					[]attr.Value{},
				)),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Tag ID. If provided, name and kind must be omitted.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("name"),
								),
							},
						},
						"name": schema.StringAttribute{
							Description: "Tag name. If provided, id must be omitted.",
							Optional:    true,
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Description: "Tag kind. If provided, name must be provided. Use this field for disambiguation when the tag name matches both a regular and a classification tag. Use 'Tag' to match a regular, user-managed tag. Use 'Classification' to match a tag that was automatically created by Sifflet. See the Sifflet documentation for more about tag types.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("Tag", "Classification"),
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("id"),
								),
							},
						},
					},
				},
			},
			"parameters": parameters.ParametersModel{}.TerraformSchema(),
		},
	}

}

func (r *sourceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SourceResourceSchema(ctx)
}

func (r *sourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 2*time.Minute)
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

	sourceResponse, err := r.client.PublicCreateSourceWithResponse(ctx, sourceDto)
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

	var newState sourceModel
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

func (r *sourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, 2*time.Minute)
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

	res, err := r.client.PublicGetSourceWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source: could not parse API response", err.Error())
		return
	}

	if res.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read source", res.StatusCode(), res.Body)
		return
	}

	var newState sourceModel
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

func (r *sourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 2*time.Minute)
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

	updateResponse, err := r.client.PublicEditSourceWithResponse(ctx, id, body)
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

	var newState sourceModel
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

func (r *sourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, 2*time.Minute)
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

	credentialResponse, _ := r.client.PublicDeleteSourceByIdWithResponse(ctx, id)

	if credentialResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete source",
			credentialResponse.StatusCode(), credentialResponse.Body,
		)
		return
	}

}

func (r *sourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *sourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r sourceResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
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
func (r *sourceResource) MoveState(ctx context.Context) []resource.StateMover {
	sourceSchema := datasource.DatasourceResourceSchema(ctx)
	return []resource.StateMover{
		{
			SourceSchema: &sourceSchema,
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				if req.SourceTypeName != "sifflet_datasource" {
					return
				}

				if req.SourceSchemaVersion != 0 {
					return
				}

				var sourceStateData datasource.CreateDatasourceDto
				resp.Diagnostics.Append(req.SourceState.Get(ctx, &sourceStateData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				tagsModel := make([]tagModel, len(*sourceStateData.Tags))
				for i, tag := range *sourceStateData.Tags {
					tagsModel[i] = tagModel{ID: tag}
				}
				tags, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: tagModel{}.AttributeTypes()}, tagsModel)
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					return
				}

				var parametersModel parameters.ParametersModel
				var timezone types.String
				if sourceStateData.BigQuery != nil {
					parametersModel, diags = parameters.BigQueryParametersModel{
						ProjectId:        sourceStateData.BigQuery.ProjectID,
						BillingProjectId: sourceStateData.BigQuery.BillingProjectID,
					}.AsParametersModel(ctx)
					timezone = moveTimezone(sourceStateData.BigQuery.TimezoneData)
				} else if sourceStateData.DBT != nil {
					parametersModel, diags = parameters.DbtParametersModel{
						ProjectName: sourceStateData.DBT.ProjectName,
						Target:      sourceStateData.DBT.Target,
					}.AsParametersModel(ctx)
					timezone = moveTimezone(sourceStateData.DBT.TimezoneData)
				} else if sourceStateData.Snowflake != nil {
					parametersModel, diags = parameters.SnowflakeParametersModel{
						AccountIdentifier: sourceStateData.Snowflake.AccountIdentifier,
						Database:          sourceStateData.Snowflake.Database,
						Schema:            sourceStateData.Snowflake.Schema,
						Warehouse:         sourceStateData.Snowflake.Warehouse,
					}.AsParametersModel(ctx)
					timezone = moveTimezone(sourceStateData.Snowflake.TimezoneData)
				} else {
					resp.Diagnostics.AddError("Unsupported source type", "The sifflet_datasource type is not supported for this move operation.")
					return
				}
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					return
				}

				parameters, diags := types.ObjectValueFrom(ctx, parametersModel.AttributeTypes(), parametersModel)
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					return
				}

				targetStateData := sourceModel{
					baseSourceModel: baseSourceModel{
						ID:          sourceStateData.ID,
						Name:        sourceStateData.Name,
						Description: types.StringNull(), // Description not available in the sifflet_datasource resource.
						Credentials: sourceStateData.SecretID,
						Schedule:    types.StringPointerValue(sourceStateData.CronExpression),
						Timezone:    timezone,
						Tags:        tags,
						Parameters:  parameters,
					},
				}

				resp.Diagnostics.Append(resp.TargetState.Set(ctx, targetStateData)...)
			},
		},
	}
}
