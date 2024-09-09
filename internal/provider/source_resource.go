package provider

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/source"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource              = &sourceResource{}
	_ resource.ResourceWithConfigure = &sourceResource{}
)

func NewSourceResource() resource.Resource {
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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the source.",
				Computed:    true,
			},
			"credential": schema.StringAttribute{
				Description: "Name of the credential used to connect to the source. Required for most datasources, except for 'athena', 'dbt' and 'quicksight' sources.",
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
				Description: "Tags for the source. For each tag, you can provider either: an ID, a name, or a name + a kind (when the name alone is ambiguous). It's recommended to use tag IDs (coming from a sifflet_tag resource or data source) most of the time, but directly providing tag names can simplify some configurations.",
				Optional:    true,
				Computed:    true,
				Default: listdefault.StaticValue(types.ListValueMust(
					types.ObjectType{
						AttrTypes: source.TagModel{}.AttributeTypes(),
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
							Description: "Tag kind. If provided, name must be provided. Use this field when a tag name could be used for different kinds of tags.",
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
			"parameters": source.ParametersModel{}.TerraformSchema(),
		},
	}

}

func (r *sourceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SourceResourceSchema(ctx)
}

func tagsModelToDto(tagsModel []source.TagModel) ([]sifflet.PublicTagReferenceDto, diag.Diagnostics) {
	tagsDto := make([]sifflet.PublicTagReferenceDto, len(tagsModel))
	for i, tagModel := range tagsModel {
		var id *uuid.UUID
		var kind *sifflet.PublicTagReferenceDtoKind
		var name *string
		if !tagModel.ID.IsNull() && tagModel.ID.ValueString() != "" {
			// If an ID was provided, the DTO should not include a name or kind
			idv, err := uuid.Parse(tagModel.ID.ValueString())
			if err != nil {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic("Tag ID is not a valid UUID", err.Error()),
				}
			}
			id = &idv
		} else {
			// If an ID is not provided, then a name was provided (enforced by the schema)
			// Let's double check that here for clarity.
			if tagModel.Name.IsNull() || tagModel.Name.ValueString() == "" {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic("Tag name is required when an ID is not provided", ""),
				}
			}
			name = tagModel.Name.ValueStringPointer()
			if !tagModel.Kind.IsNull() && tagModel.Kind.ValueString() != "" {
				kindv, err := source.ParseTagKind(tagModel.Kind.ValueString())
				if err != nil {
					return nil, diag.Diagnostics{
						diag.NewErrorDiagnostic("Could not parse provided tag kind", err.Error()),
					}
				}
				kind = &kindv
			}
		}
		tagsDto[i] = sifflet.PublicTagReferenceDto{
			Id:   id,
			Name: name,
			Kind: kind,
		}
	}
	return tagsDto, nil
}

func tagsDtoToModel(tagsDto []sifflet.PublicTagReferenceDto) ([]source.TagModel, diag.Diagnostics) {
	tagsModel := make([]source.TagModel, len(tagsDto))
	for i, tagDto := range tagsDto {
		kind, err := source.TagKindToString(*tagDto.Kind)
		if err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to read source: could not parse tag kind", err.Error()),
			}
		}
		tagModel := source.TagModel{
			ID:   types.StringValue(tagDto.Id.String()),
			Name: types.StringPointerValue(tagDto.Name),
			Kind: types.StringValue(kind),
		}
		tagsModel[i] = tagModel
	}
	return tagsModel, nil
}

func (r *sourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan source.SourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var parametersModel source.ParametersModel
	resp.Diagnostics.Append(
		plan.Parameters.As(ctx, &parametersModel, basetypes.ObjectAsOptions{})...,
	)
	if resp.Diagnostics.HasError() {
		return
	}
	// Deduce the source type from the parameters
	// We need to do this now, since later code can rely on the source type being set in the model.
	err := parametersModel.SetSourceType()
	if err != nil {
		resp.Diagnostics.AddError("Unable to create source", err.Error())
		return
	}

	t, err := parametersModel.GetSourceType()
	if err != nil {
		resp.Diagnostics.AddError("Unable to create source", err.Error())
		return
	}
	var credentials *string
	if t.RequiresCredential() {
		credentials = plan.Credential.ValueStringPointer()
		if credentials == nil {
			resp.Diagnostics.AddError("Unable to create source", "Credential is required for this source type, but got an empty string")
			return
		}
	} else {
		credentials = nil
		if !plan.Credential.IsNull() {
			resp.Diagnostics.AddError("Invalid credential", "Credential is not required for this source type and would be ignored, but got a non-null string")
			return
		}
	}

	parametersDto, diags := parametersModel.AsDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	length := len(plan.Tags.Elements())
	tagsModel := make([]source.TagModel, 0, length)
	diags = plan.Tags.ElementsAs(ctx, &tagsModel, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagsDto, diags := tagsModelToDto(tagsModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialDto := sifflet.PublicCreateSourceDto{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Credentials: credentials,
		Schedule:    plan.Schedule.ValueStringPointer(),
		Timezone:    plan.Timezone.ValueStringPointer(),
		Parameters:  parametersDto,
		Tags:        &tagsDto,
	}

	sourceResponse, err := r.client.PublicCreateSourceWithResponse(ctx, credentialDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create source", err.Error())
		return
	}

	if sourceResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create source", sourceResponse.StatusCode(), sourceResponse.Body,
		)
		resp.State.RemoveResource(ctx)
		return
	}

	tagsModel, diags = tagsDtoToModel(*sourceResponse.JSON201.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tags, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: source.TagModel{}.AttributeTypes()}, tagsModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newStateModel := source.SourceModel{
		ID:          types.StringValue(sourceResponse.JSON201.Id.String()),
		Name:        types.StringValue(sourceResponse.JSON201.Name),
		Description: types.StringPointerValue(sourceResponse.JSON201.Description),
		Credential:  types.StringPointerValue(sourceResponse.JSON201.Credentials),
		Schedule:    types.StringPointerValue(sourceResponse.JSON201.Schedule),
		Timezone:    types.StringPointerValue(sourceResponse.JSON201.Timezone),
		Parameters:  plan.Parameters,
		Tags:        tags,
	}
	newState, diags := types.ObjectValueFrom(ctx, newStateModel.AttributeTypes(), newStateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.SetAttribute(ctx, path.Root("parameters").AtName("source_type"), parametersModel.SourceType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state source.SourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source: could not parse source ID stored in state as an UUID", err.Error())
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
		resp.State.RemoveResource(ctx)
		return
	}

	sourceType, diags := source.ParseSourceType(res)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceTypeParams, err := source.ParamsImplFromApiResponseName(sourceType)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source: unsupported source type", err.Error())
		return
	}
	diags = sourceTypeParams.ModelFromDto(ctx, res.JSON200.Parameters)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	parametersModel, diags := sourceTypeParams.AsParametersModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	parameters, diags := types.ObjectValueFrom(
		ctx,
		parametersModel.AttributeTypes(),
		parametersModel,
	)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	tagModels := make([]source.TagModel, len(*res.JSON200.Tags))
	for i, tag := range *res.JSON200.Tags {
		kind, err := source.TagKindToString(*tag.Kind)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read source: could not parse tag kind", err.Error())
			return
		}
		tagModel := source.TagModel{
			ID:   types.StringValue(tag.Id.String()),
			Name: types.StringPointerValue(tag.Name),
			Kind: types.StringValue(kind),
		}
		tagModels[i] = tagModel
	}

	tags, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: source.TagModel{}.AttributeTypes()}, tagModels)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	state = source.SourceModel{
		ID:          types.StringValue(res.JSON200.Id.String()),
		Name:        types.StringValue(res.JSON200.Name),
		Description: types.StringPointerValue(res.JSON200.Description),
		Credential:  types.StringPointerValue(res.JSON200.Credentials),
		Schedule:    types.StringPointerValue(res.JSON200.Schedule),
		Timezone:    types.StringPointerValue(res.JSON200.Timezone),
		Parameters:  parameters,
		Tags:        tags,
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// TODO: factorize duplicated code with Create.
func (r *sourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan source.SourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update source: unable to parse state source ID as uuid", err.Error())
		return
	}

	var parametersModel source.ParametersModel
	resp.Diagnostics.Append(
		plan.Parameters.As(ctx, &parametersModel, basetypes.ObjectAsOptions{})...,
	)
	if resp.Diagnostics.HasError() {
		return
	}
	// Deduce the source type from the parameters
	// We need to do this now, since later code can rely on the source type being set in the model.
	err = parametersModel.SetSourceType()
	if err != nil {
		resp.Diagnostics.AddError("Unable to update source", err.Error())
		return
	}

	t, err := parametersModel.GetSourceType()
	if err != nil {
		resp.Diagnostics.AddError("Unable to update source", err.Error())
		return
	}

	var credentials *string
	if t.RequiresCredential() {
		credentials = plan.Credential.ValueStringPointer()
		if credentials == nil {
			resp.Diagnostics.AddError("Unable to create source", "Credential is required for this source type, but got an empty string")
			return
		}
	} else {
		credentials = nil
		if !plan.Credential.IsNull() {
			resp.Diagnostics.AddError("Invalid credential", "Credential is not required for this source type and would be ignored, but got a non-null string")
			return
		}
	}

	length := len(plan.Tags.Elements())
	tagsModel := make([]source.TagModel, 0, length)
	diags = plan.Tags.ElementsAs(ctx, &tagsModel, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagsDto, diags := tagsModelToDto(tagsModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := sifflet.PublicEditSourceJSONRequestBody{
		Description: plan.Description.ValueStringPointer(),
		Credentials: credentials,
		Schedule:    plan.Schedule.ValueStringPointer(),
		Timezone:    plan.Timezone.ValueStringPointer(),
		Name:        plan.Name.ValueStringPointer(),
		Tags:        &tagsDto,
		// TODO: add support for parameters - the API documentation doesn't explain how to update them yet. See PLTE-964.
		// For now, parameter changes are marked as "require replacement" in the schema.
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

	tagsModel, diags = tagsDtoToModel(*updateResponse.JSON200.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tags, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: source.TagModel{}.AttributeTypes()}, tagsModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newStateModel := source.SourceModel{
		ID:          types.StringValue(updateResponse.JSON200.Id.String()),
		Name:        types.StringValue(updateResponse.JSON200.Name),
		Description: types.StringPointerValue(updateResponse.JSON200.Description),
		Credential:  types.StringPointerValue(updateResponse.JSON200.Credentials),
		Schedule:    types.StringPointerValue(updateResponse.JSON200.Schedule),
		Timezone:    types.StringPointerValue(updateResponse.JSON200.Timezone),
		// Copying the plan parameters since any change will require a replacement (see PLTE-964)
		Parameters: plan.Parameters,
		Tags:       tags,
	}
	newState, diags := types.ObjectValueFrom(ctx, newStateModel.AttributeTypes(), newStateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.SetAttribute(ctx, path.Root("parameters").AtName("source_type"), parametersModel.SourceType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state source.SourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read source", err.Error())
		return
	}

	credentialResponse, _ := r.client.PublicDeleteSourceByIdWithResponse(ctx, id)

	if credentialResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete source",
			credentialResponse.StatusCode(), credentialResponse.Body,
		)
		resp.State.RemoveResource(ctx)
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
	types := source.GetAllSourceTypes()
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
