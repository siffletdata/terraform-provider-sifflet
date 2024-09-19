package source

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type AirflowParametersModel struct {
	Host types.String `tfsdk:"host"`
	Port types.Int32  `tfsdk:"port"`
}

func (m AirflowParametersModel) SchemaSourceType() string {
	return "airflow"
}

func (m AirflowParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Airflow API host",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Airflow API port",
				Required:    true,
			},
		},
	}
}

func (m AirflowParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host": types.StringType,
		"port": types.Int32Type,
	}
}

func (m AirflowParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	airflowParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Airflow = airflowParams
	return o, diag.Diagnostics{}
}

func (m AirflowParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Airflow.IsNull()
}

func (m *AirflowParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Airflow.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicAirflowParametersDto{
		Host: m.Host.ValueStringPointer(),
		Port: m.Port.ValueInt32Pointer(),
		Type: sifflet.PublicAirflowParametersDtoTypeAIRFLOW,
	}
	err := parametersDto.FromPublicAirflowParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *AirflowParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Airflow.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicAirflowParametersDto{
		Host: m.Host.ValueStringPointer(),
		Port: m.Port.ValueInt32Pointer(),
		Type: sifflet.PublicAirflowParametersDtoTypeAIRFLOW,
	}
	err := parametersDto.FromPublicAirflowParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *AirflowParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicAirflowParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	return diag.Diagnostics{}
}

func (m AirflowParametersModel) RequiresCredential() bool {
	return true
}
