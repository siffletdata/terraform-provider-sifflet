package source_v2

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	parameters "terraform-provider-sifflet/internal/provider/source_v2/parameters_v2"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource               = &sourceV2Resource{}
	_ resource.ResourceWithConfigure  = &sourceV2Resource{}
	_ resource.ResourceWithModifyPlan = &sourceV2Resource{}
)

// ModifyPlan sets the computed source_type attribute based on the parameters
// This allows users to see the detected source type in terraform plan output
// before the resource is created or updated.
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

	var parametersModel parameters.ParametersModel
	resp.Diagnostics.Append(plan.Parameters.As(ctx, &parametersModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceType, diags := parametersModel.GetSourceParameter(ctx)
	if diags.HasError() {
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

func SourceV2ResourceSchema(ctx context.Context) schema.Schema {
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
			"parameters": parameters.ParametersModel{}.TerraformSchema(),
		},
	}

}

func (r *sourceV2Resource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SourceV2ResourceSchema(ctx)
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

	sourceBody, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := sifflet.PublicCreateSourceV2JSONRequestBody(sourceBody)
	sourceResponse, err := r.client.PublicCreateSourceV2WithResponse(ctx, requestBody)
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

	var newState sourceV2Model
	var sourceDto sifflet.SiffletPublicGetSourceV2Dto
	err = sourceDto.UnmarshalJSON(sourceResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source", err.Error())
	}
	diags = newState.FromDto(ctx, sourceDto)
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

	var newState sourceV2Model
	var sourceDto sifflet.SiffletPublicGetSourceV2Dto
	err = sourceDto.UnmarshalJSON(res.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source", err.Error())
	}
	diags = newState.FromDto(ctx, sourceDto)
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
	updateResponse, err := r.client.PublicEditSourceV2WithResponse(ctx, id, sifflet.PublicEditSourceV2JSONRequestBody(body))
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

	var newState sourceV2Model
	var sourceDto sifflet.SiffletPublicGetSourceV2Dto
	err = sourceDto.UnmarshalJSON(updateResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source", err.Error())
	}
	diags = newState.FromDto(ctx, sourceDto)
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

	deleteResponse, err := r.client.PublicDeleteSourceV2WithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete source", err.Error())
		return
	}

	if deleteResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete source",
			deleteResponse.StatusCode(), deleteResponse.Body,
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

// ConfigValidators ensures exactly one source type is configured in parameters
// This prevents users from accidentally configuring multiple source types
// which would result in ambiguous behavior.
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
