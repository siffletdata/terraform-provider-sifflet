package source

import (
	"context"
	"terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/provider/source/parameters"
	"terraform-provider-sifflet/internal/tfutils"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type sourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
	Timezone    types.String `tfsdk:"timezone"`
	Parameters  types.Object `tfsdk:"parameters"`
	Tags        types.List   `tfsdk:"tags"`
}

var (
	_ model.FullModel[sifflet.PublicGetSourceDto, sifflet.PublicCreateSourceDto, sifflet.PublicEditSourceJSONRequestBody] = &sourceModel{}
	_ model.ModelWithId[uuid.UUID]                                                                                        = sourceModel{}
)

func (m sourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
		"timezone":    types.StringType,
		"parameters": types.ObjectType{
			AttrTypes: parameters.ParametersModel{}.AttributeTypes(),
		},
		"tags": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: tagModel{}.AttributeTypes(),
			},
		},
	}
}

func (m sourceModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.ID.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

func (m sourceModel) getTags() ([]tagModel, diag.Diagnostics) {
	tags := make([]tagModel, 0, len(m.Tags.Elements()))
	diags := m.Tags.ElementsAs(context.Background(), &tags, false)
	return tags, diags
}

func (m sourceModel) getParameters(ctx context.Context) (parameters.ParametersModel, diag.Diagnostics) {
	var parametersModel parameters.ParametersModel
	diags := m.Parameters.As(ctx, &parametersModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return parametersModel, diags
	}
	// Deduce the source type from the parameters
	// We need to do this now, since later code can rely on the source type being set in the model.
	err := parametersModel.SetSourceType()
	if err != nil {
		return parametersModel, tfutils.ErrToDiags("Unsupported source type", err)
	}
	return parametersModel, diags
}

func (m sourceModel) getCredentialsName(ctx context.Context) (*string, diag.Diagnostics) {
	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return nil, diags
	}

	sourceType, err := parametersModel.GetSourceType()
	if err != nil {
		return nil, tfutils.ErrToDiags("Unsupported source type", err)
	}

	var credentials *string
	if sourceType.RequiresCredential() {
		credentials = m.Credentials.ValueStringPointer()
		if credentials == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Invalid credentials", "Credentials are required for this source type, but got an empty string"),
			}
		}
	} else {
		credentials = nil
		if !m.Credentials.IsNull() {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Invalid credentials", "Credentials are not required for this source type and would be ignored, but got a non-null string"),
			}
		}
	}

	return credentials, diag.Diagnostics{}
}

func parametersDtoToModel(ctx context.Context, dto sifflet.PublicGetSourceDto_Parameters) (parameters.ParametersModel, diag.Diagnostics) {
	sourceType, err := sifflet.GetSourceType(dto)
	if err != nil {
		return parameters.ParametersModel{}, tfutils.ErrToDiags("Unable to read source", err)
	}
	sourceTypeParams, err := parameters.ParamsImplFromApiResponseName(sourceType)
	if err != nil {
		return parameters.ParametersModel{}, tfutils.ErrToDiags("Unsupported source type", err)
	}
	diags := sourceTypeParams.ModelFromDto(ctx, dto)
	if diags.HasError() {
		return parameters.ParametersModel{}, diags
	}
	out, diags := sourceTypeParams.AsParametersModel(ctx)
	if diags.HasError() {
		return parameters.ParametersModel{}, diags
	}
	out.SourceType = types.StringValue(sourceTypeParams.SchemaSourceType())
	return out, diag.Diagnostics{}
}

func (m *sourceModel) FromDto(ctx context.Context, dto sifflet.PublicGetSourceDto) diag.Diagnostics {
	tags, diags := model.NewModelListFromDto(ctx, *dto.Tags,
		func() model.InnerModel[client.PublicTagReferenceDto] { return &tagModel{} },
	)
	if diags.HasError() {
		return diags
	}

	parametersModel, diags := parametersDtoToModel(ctx, dto.Parameters)
	if diags.HasError() {
		return diags
	}
	parameters, diags := types.ObjectValueFrom(ctx, parameters.ParametersModel{}.AttributeTypes(), parametersModel)
	if diags.HasError() {
		return diags
	}

	m.ID = types.StringValue(dto.Id.String())
	m.Name = types.StringValue(dto.Name)
	m.Description = types.StringPointerValue(dto.Description)
	m.Credentials = types.StringPointerValue(dto.Credentials)
	m.Schedule = types.StringPointerValue(dto.Schedule)
	m.Timezone = types.StringPointerValue(dto.Timezone)
	m.Parameters = parameters
	m.Tags = tags
	return diag.Diagnostics{}
}

func (m sourceModel) ToCreateDto(ctx context.Context) (sifflet.PublicCreateSourceDto, diag.Diagnostics) {
	tagsModel, diags := m.getTags()
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto{}, diags
	}

	tagsDto, diags := tfutils.MapWithDiagnostics(
		tagsModel,
		func(tagModel tagModel) (sifflet.PublicTagReferenceDto, diag.Diagnostics) {
			return tagModel.ToDto()
		},
	)
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto{}, diags
	}

	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto{}, diags
	}

	credentialsName, diags := m.getCredentialsName(ctx)
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto{}, diags
	}

	parametersDto, diags := parametersModel.AsCreateSourceDto(ctx)
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto{}, diags
	}

	return sifflet.PublicCreateSourceDto{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
		Credentials: credentialsName,
		Schedule:    m.Schedule.ValueStringPointer(),
		Timezone:    m.Timezone.ValueStringPointer(),
		Parameters:  parametersDto,
		Tags:        &tagsDto,
	}, diag.Diagnostics{}
}

func (m sourceModel) ToUpdateDto(ctx context.Context) (sifflet.PublicEditSourceJSONRequestBody, diag.Diagnostics) {
	credentialsName, diags := m.getCredentialsName(ctx)
	if diags.HasError() {
		return sifflet.PublicEditSourceJSONRequestBody{}, diags
	}

	tagsModel, diags := m.getTags()
	if diags.HasError() {
		return sifflet.PublicEditSourceJSONRequestBody{}, diags
	}
	tagsDto, diags := tfutils.MapWithDiagnostics(tagsModel,
		func(tagModel tagModel) (sifflet.PublicTagReferenceDto, diag.Diagnostics) {
			return tagModel.ToDto()
		},
	)
	if diags.HasError() {
		return sifflet.PublicEditSourceJSONRequestBody{}, diags
	}

	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return sifflet.PublicEditSourceJSONRequestBody{}, diags
	}
	parametersDto, diags := parametersModel.AsUpdateSourceDto(ctx)
	if diags.HasError() {
		return sifflet.PublicEditSourceJSONRequestBody{}, diags
	}

	return sifflet.PublicEditSourceJSONRequestBody{
		Description: m.Description.ValueStringPointer(),
		Credentials: credentialsName,
		Schedule:    m.Schedule.ValueStringPointer(),
		Timezone:    m.Timezone.ValueStringPointer(),
		Name:        m.Name.ValueStringPointer(),
		Tags:        &tagsDto,
		Parameters:  &parametersDto,
	}, diag.Diagnostics{}
}

type tagModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
	Kind types.String `tfsdk:"kind"`
}

var (
	_ model.InnerModel[sifflet.PublicTagReferenceDto] = &tagModel{}
)

func (m tagModel) ToDto() (sifflet.PublicTagReferenceDto, diag.Diagnostics) {
	var id *uuid.UUID
	var kind *sifflet.PublicTagReferenceDtoKind
	var name *string
	if !m.ID.IsNull() && m.ID.ValueString() != "" {
		// If an ID was provided, the DTO should not include a name or kind
		idv, err := uuid.Parse(m.ID.ValueString())
		if err != nil {
			return sifflet.PublicTagReferenceDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Tag ID is not a valid UUID", err.Error()),
			}
		}
		id = &idv
	} else {
		// If an ID is not provided, then a name was provided (enforced by the schema)
		// Let's double check that here for clarity.
		if m.Name.IsNull() || m.Name.ValueString() == "" {
			return sifflet.PublicTagReferenceDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Tag name is required when an ID is not provided", ""),
			}
		}
		name = m.Name.ValueStringPointer()
		if !m.Kind.IsNull() && m.Kind.ValueString() != "" {
			t := sifflet.PublicTagReferenceDtoKind(m.Kind.ValueString())
			kind = &t
		}
	}

	return sifflet.PublicTagReferenceDto{
		Id:   id,
		Name: name,
		Kind: kind,
	}, diag.Diagnostics{}
}

func (m *tagModel) FromDto(_ context.Context, dto sifflet.PublicTagReferenceDto) diag.Diagnostics {
	m.ID = types.StringValue(dto.Id.String())
	m.Name = types.StringPointerValue(dto.Name)
	kind := "Tag"
	if dto.Kind != nil {
		kind = string(*dto.Kind)
	}
	m.Kind = types.StringValue(kind)
	return diag.Diagnostics{}
}

func (m tagModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
		"kind": types.StringType,
	}
}
