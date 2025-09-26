package source_v2

import (
	"context"
	"terraform-provider-sifflet/internal/model"
	parameters "terraform-provider-sifflet/internal/provider/source_v2/parameters_v2"
	"terraform-provider-sifflet/internal/tfutils"

	"terraform-provider-sifflet/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type baseSourceV2Model struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Timezone   types.String `tfsdk:"timezone"`
	Parameters types.Object `tfsdk:"parameters"`
}

type sourceV2Model struct {
	baseSourceV2Model
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ model.FullModel[client.SiffletPublicGetSourceV2Dto, client.PublicCreateSourceV2JSONBody, client.PublicEditSourceV2JSONBody] = &sourceV2Model{}
	_ model.ModelWithId[uuid.UUID]                                                                                                = sourceV2Model{}
)

func (m baseSourceV2Model) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":       types.StringType,
		"name":     types.StringType,
		"timezone": types.StringType,
		"parameters": types.ObjectType{
			AttrTypes: parameters.ParametersModel{}.AttributeTypes(),
		},
	}
}

func (m sourceV2Model) AttributeTypes() map[string]attr.Type {
	attrs := m.baseSourceV2Model.AttributeTypes()
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

func (m sourceV2Model) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.ID.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

func parametersModelFromGetSourceV2Dto(ctx context.Context, dto client.SiffletPublicGetSourceV2Dto) (parameters.ParametersModel, diag.Diagnostics) {
	sourceParams, diags := parameters.SourceParametersModelFromDto(ctx, dto)
	if diags.HasError() {
		return parameters.ParametersModel{}, diags
	}
	out, diags := sourceParams.AsParametersModel(ctx)
	if diags.HasError() {
		return parameters.ParametersModel{}, diags
	}
	out.SourceType = types.StringValue(sourceParams.SchemaSourceType())
	return out, diag.Diagnostics{}
}

func (m *baseSourceV2Model) FromDto(ctx context.Context, dto client.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	parametersModel, diags := parametersModelFromGetSourceV2Dto(ctx, dto)
	if diags.HasError() {
		return diags
	}
	parameters, diags := types.ObjectValueFrom(ctx, parameters.ParametersModel{}.AttributeTypes(), parametersModel)
	if diags.HasError() {
		return diags
	}
	m.Parameters = parameters

	sourceDto, err := dto.GetSourceDto()
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read source", err.Error())}
	}
	m.ID = types.StringValue(sourceDto.GetId().String())
	m.Name = types.StringValue(sourceDto.GetName())
	m.Timezone = types.StringPointerValue(sourceDto.GetTimezone())

	return diag.Diagnostics{}
}

func (m *sourceV2Model) FromDto(ctx context.Context, dto client.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	return m.baseSourceV2Model.FromDto(ctx, dto)
}

func (m sourceV2Model) getParameters(ctx context.Context) (parameters.ParametersModel, diag.Diagnostics) {
	var parametersModel parameters.ParametersModel
	diags := m.Parameters.As(ctx, &parametersModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return parametersModel, diags
	}
	// Deduce the source type from the parameters
	// We need to do this now, since later code can rely on the source type being set in the model.
	diags = parametersModel.SetSourceType(ctx)
	if diags.HasError() {
		return parametersModel, diags
	}
	return parametersModel, diags
}

func (m sourceV2Model) ToCreateDto(ctx context.Context) (client.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return client.PublicCreateSourceV2JSONBody{}, diags
	}

	createDto, diags := parametersModel.ToCreateDto(ctx, m.Name.ValueString(), m.Timezone.ValueString())

	return createDto, diags
}

func (m sourceV2Model) ToUpdateDto(ctx context.Context) (client.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	parametersModel, diags := m.getParameters(ctx)
	if diags.HasError() {
		return client.PublicEditSourceV2JSONBody{}, diags
	}

	updateDto, diags := parametersModel.ToUpdateDto(ctx, m.Name.ValueString(), m.Timezone.ValueString())

	return updateDto, diags
}
