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

type DatabricksParametersModel struct {
	Catalog  types.String `tfsdk:"catalog"`
	Host     types.String `tfsdk:"host"`
	HttpPath types.String `tfsdk:"http_path"`
	Port     types.Int32  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
}

func (m DatabricksParametersModel) SchemaSourceType() string {
	return "databricks"
}

func (m DatabricksParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"catalog": schema.StringAttribute{
				Description: "Databricks catalog name",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "Databricks host",
				Required:    true,
			},
			"http_path": schema.StringAttribute{
				Description: "Databricks HTTP path",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Databricks server port",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Databricks schema",
				Required:    true,
			},
		},
	}
}

func (m DatabricksParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"catalog":   types.StringType,
		"host":      types.StringType,
		"http_path": types.StringType,
		"port":      types.Int32Type,
		"schema":    types.StringType,
	}
}

func (m DatabricksParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	databricksParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Databricks = databricksParams
	return o, diag.Diagnostics{}
}

func (m DatabricksParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Databricks.IsNull()
}

func (m *DatabricksParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Databricks.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicDatabricksParametersDto{
		Type:     sifflet.PublicDatabricksParametersDtoTypeDATABRICKS,
		Catalog:  m.Catalog.ValueStringPointer(),
		Host:     m.Host.ValueStringPointer(),
		HttpPath: m.HttpPath.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
	}
	err := parametersDto.FromPublicDatabricksParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *DatabricksParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicDatabricksParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Catalog = types.StringPointerValue(paramsDto.Catalog)
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.HttpPath = types.StringPointerValue(paramsDto.HttpPath)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	m.Schema = types.StringPointerValue(paramsDto.Schema)
	return diag.Diagnostics{}
}

func (m DatabricksParametersModel) RequiresCredential() bool {
	return true
}
