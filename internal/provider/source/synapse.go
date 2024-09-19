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

type SynapseParametersModel struct {
	Database types.String `tfsdk:"database"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
}

func (m SynapseParametersModel) SchemaSourceType() string {
	return "synapse"
}

func (m SynapseParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Azure Synapse server hostname",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Azure Synapse server port number",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Schema name",
				Required:    true,
			},
		},
	}
}

func (m SynapseParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"database": types.StringType,
		"host":     types.StringType,
		"port":     types.Int32Type,
		"schema":   types.StringType,
	}
}

func (m SynapseParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	synapseParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Synapse = synapseParams
	return o, diag.Diagnostics{}
}

func (m SynapseParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Synapse.IsNull()
}

func (m *SynapseParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Synapse.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicSynapseParametersDto{
		Type:     sifflet.PublicSynapseParametersDtoTypeSYNAPSE,
		Database: m.Database.ValueStringPointer(),
		Host:     m.Host.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
	}
	err := parametersDto.FromPublicSynapseParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *SynapseParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Synapse.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicSynapseParametersDto{
		Type:     sifflet.PublicSynapseParametersDtoTypeSYNAPSE,
		Database: m.Database.ValueStringPointer(),
		Host:     m.Host.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
	}
	err := parametersDto.FromPublicSynapseParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *SynapseParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicSynapseParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	m.Schema = types.StringPointerValue(paramsDto.Schema)
	return diag.Diagnostics{}
}

func (m SynapseParametersModel) RequiresCredential() bool {
	return true
}
