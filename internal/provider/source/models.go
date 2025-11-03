package source

import (
	"context"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/provider/source/parameters"
	"terraform-provider-sifflet/internal/provider/tag"
	"terraform-provider-sifflet/internal/tfutils"

	"terraform-provider-sifflet/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type baseSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
	Timezone    types.String `tfsdk:"timezone"`
	Parameters  types.Object `tfsdk:"parameters"`
	Tags        types.List   `tfsdk:"tags"`
}

type sourceModel struct {
	baseSourceModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ model.FullModel[client.PublicGetSourceDto, client.PublicCreateSourceDto, client.PublicEditSourceJSONRequestBody] = &sourceModel{}
	_ model.ModelWithId[uuid.UUID]                                                                                     = sourceModel{}
	_ model.ReadableModel[client.PublicGetSourceDto]                                                                   = &baseSourceModel{}
)

func (m baseSourceModel) AttributeTypes() map[string]attr.Type {
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
				AttrTypes: tag.PublicApiTagModel{}.AttributeTypes(),
			},
		},
	}
}

func (m sourceModel) AttributeTypes() map[string]attr.Type {
	attrs := m.baseSourceModel.AttributeTypes()
	attrs["timeouts"] = timeouts.Type{
		ObjectType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"create": types.StringType,
				"read":   types.StringType,
				"update": types.StringType,
				"delete": types.StringType,
			},
		},
	}
	return attrs
}

func (m sourceModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.ID.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

func (m sourceModel) getTags() ([]tag.PublicApiTagModel, diag.Diagnostics) {
	tags := make([]tag.PublicApiTagModel, 0, len(m.Tags.Elements()))
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

func parametersDtoToModel(ctx context.Context, dto client.PublicGetSourceDto_Parameters) (parameters.ParametersModel, diag.Diagnostics) {
	sourceType, err := client.GetSourceType(dto)
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

func (m *baseSourceModel) FromDto(ctx context.Context, dto client.PublicGetSourceDto) diag.Diagnostics {
	tags, diags := model.NewModelListFromDto(ctx, *dto.Tags,
		func() model.InnerModel[client.PublicTagReferenceDto] { return &tag.PublicApiTagModel{} },
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

func (m *sourceModel) FromDto(ctx context.Context, dto client.PublicGetSourceDto) diag.Diagnostics {
	return m.baseSourceModel.FromDto(ctx, dto)
}

func (m sourceModel) ToCreateDto(ctx context.Context) (client.PublicCreateSourceDto, diag.Diagnostics) {
	tagsModel, diags := m.getTags()
	if diags.HasError() {
		return client.PublicCreateSourceDto{}, diags
	}

	tagsDto, diags := tfutils.MapWithDiagnostics(
		tagsModel,
		func(tagModel tag.PublicApiTagModel) (client.PublicTagReferenceDto, diag.Diagnostics) {
			return tagModel.ToDto(ctx)
		},
	)
	if diags.HasError() {
		return client.PublicCreateSourceDto{}, diags
	}

	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return client.PublicCreateSourceDto{}, diags
	}

	credentialsName, diags := m.getCredentialsName(ctx)
	if diags.HasError() {
		return client.PublicCreateSourceDto{}, diags
	}

	parametersDto, diags := parametersModel.AsCreateSourceDto(ctx)
	if diags.HasError() {
		return client.PublicCreateSourceDto{}, diags
	}

	return client.PublicCreateSourceDto{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
		Credentials: credentialsName,
		Schedule:    m.Schedule.ValueStringPointer(),
		Timezone:    m.Timezone.ValueStringPointer(),
		Parameters:  parametersDto,
		Tags:        &tagsDto,
	}, diag.Diagnostics{}
}

func (m sourceModel) ToUpdateDto(ctx context.Context) (client.PublicEditSourceJSONRequestBody, diag.Diagnostics) {
	credentialsName, diags := m.getCredentialsName(ctx)
	if diags.HasError() {
		return client.PublicEditSourceJSONRequestBody{}, diags
	}

	tagsModel, diags := m.getTags()
	if diags.HasError() {
		return client.PublicEditSourceJSONRequestBody{}, diags
	}
	tagsDto, diags := tfutils.MapWithDiagnostics(tagsModel,
		func(tagModel tag.PublicApiTagModel) (client.PublicTagReferenceDto, diag.Diagnostics) {
			return tagModel.ToDto(ctx)
		},
	)
	if diags.HasError() {
		return client.PublicEditSourceJSONRequestBody{}, diags
	}

	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return client.PublicEditSourceJSONRequestBody{}, diags
	}
	parametersDto, diags := parametersModel.AsUpdateSourceDto(ctx)
	if diags.HasError() {
		return client.PublicEditSourceJSONRequestBody{}, diags
	}

	return client.PublicEditSourceJSONRequestBody{
		Description: m.Description.ValueStringPointer(),
		Credentials: credentialsName,
		Schedule:    m.Schedule.ValueStringPointer(),
		Timezone:    m.Timezone.ValueStringPointer(),
		Name:        m.Name.ValueStringPointer(),
		Tags:        &tagsDto,
		Parameters:  &parametersDto,
	}, diag.Diagnostics{}
}
