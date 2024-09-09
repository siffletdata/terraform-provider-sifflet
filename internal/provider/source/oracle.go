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

type OracleParametersModel struct {
	Host     types.String `tfsdk:"host"`
	Database types.String `tfsdk:"database"`
	Port     types.Int32  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
}

func (m OracleParametersModel) SchemaSourceType() string {
	return "oracle"
}

func (m OracleParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Oracle server hostname",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Oracle server port number",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Schema name",
				Required:    true,
			},
		},
	}
}

func (m OracleParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":     types.StringType,
		"database": types.StringType,
		"port":     types.Int32Type,
		"schema":   types.StringType,
	}
}

func (m OracleParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	oracleParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Oracle = oracleParams
	return o, diag.Diagnostics{}
}

func (m OracleParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Oracle.IsNull()
}

func (m *OracleParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Oracle.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicOracleParametersDto{
		Type:     sifflet.PublicOracleParametersDtoTypeORACLE,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
	}
	err := parametersDto.FromPublicOracleParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *OracleParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicOracleParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	m.Schema = types.StringPointerValue(paramsDto.Schema)
	return diag.Diagnostics{}
}

func (m OracleParametersModel) RequiresCredential() bool {
	return true
}
