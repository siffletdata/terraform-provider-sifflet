package source

import (
	"context"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/provider/source/parameters"
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
	// UNKNOWN: PublicUpdateSourceV2Dto doesn't exist because there isn't an allOf on the PublicUpdate<type>SourceV2Dtos, see how to fix that
	_ model.FullModel[client.PublicGetSourceV2Dto, client.PublicCreateSourceV2Dto, client.PublicUpdateSourceV2Dto] = &sourceV2Model{}
	_ model.ModelWithId[uuid.UUID]                                                                                 = sourceV2Model{}
	_ model.ReadableModel[client.PublicGetSourceV2Dto]                                                             = &baseSourceV2Model{}
)

func (m baseSourceV2Model) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":       types.StringType,
		"name":     types.StringType,
		"timezone": types.StringType,
		"parameters": types.ObjectType{
			// TODO: change to v2 parameters model
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

func (m sourceV2Model) getParameters(ctx context.Context) (parameters.ParametersModel, diag.Diagnostics) {
	// TODO: change to v2 parameters model
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

func parametersDtoToModel(ctx context.Context, dto client.PublicGetSourceV2Dto) (parameters.ParametersModel, diag.Diagnostics) {
	sourceType := dto.Type
	// TODO: change to v2 parameters model
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

func (m *baseSourceV2Model) FromDto(ctx context.Context, dto client.PublicGetSourceV2Dto) diag.Diagnostics {
	parametersModel, diags := parametersDtoToModel(ctx, dto)
	if diags.HasError() {
		return diags
	}
	parameters, diags := types.ObjectValueFrom(ctx, parameters.ParametersModel{}.AttributeTypes(), parametersModel)
	if diags.HasError() {
		return diags
	}

	m.ID = types.StringValue(dto.Id.String())
	m.Name = types.StringValue(dto.Name)
	m.Timezone = types.StringPointerValue(dto.Timezone)
	m.Parameters = parameters
	return diag.Diagnostics{}
}

func (m *sourceV2Model) FromDto(ctx context.Context, dto client.PublicGetSourceV2Dto) diag.Diagnostics {
	return m.baseSourceV2Model.FromDto(ctx, dto)
}

func (m sourceV2Model) ToCreateDto(ctx context.Context) (client.PublicCreateSourceV2Dto, diag.Diagnostics) {
	// TODO: change to v2 parameters model
	// parametersModel, diags := m.getParameters(ctx)
	// if diags.HasError() {
	// 	return client.PublicCreateSourceV2Dto{}, diags
	// }

	// TODO: create DTO with the help of the parameters model
	return client.PublicCreateSourceV2Dto{
		Name:     m.Name.ValueString(),
		Timezone: m.Timezone.ValueStringPointer(),
		// Type:     parametersModel.SourceType,
	}, diag.Diagnostics{}
}

// UNKNOWN: PublicUpdateSourceV2Dto doesn't exist because there isn't an allOf on the PublicUpdate<type>SourceV2Dtos, see how to fix that
func (m sourceV2Model) ToUpdateDto(ctx context.Context) (client.PublicUpdateSourceV2Dto, diag.Diagnostics) {
	// TODO: change to v2 parameters model
	// parametersModel, diags := m.getParameters(ctx)
	// if diags.HasError() {
	// 	return client.PublicEditSourceV2JSONRequestBody{}, diags
	// }

	// TODO: create DTO with the help of the parameters model
	return client.PublicUpdateSourceV2Dto{
		Timezone: m.Timezone.ValueStringPointer(),
		Name:     m.Name.ValueStringPointer(),
		// Type:     parametersModel.SourceType,
	}, diag.Diagnostics{}
}
