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

type PostgresqlParametersModel struct {
	Host     types.String `tfsdk:"host"`
	Database types.String `tfsdk:"database"`
	Port     types.Int32  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
}

func (m PostgresqlParametersModel) SchemaSourceType() string {
	return "postgresql"
}

func (m PostgresqlParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "PostgreSQL server hostname",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "PostgreSQL server port number",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Schema name",
				Required:    true,
			},
		},
	}
}

func (m PostgresqlParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":     types.StringType,
		"database": types.StringType,
		"port":     types.Int32Type,
		"schema":   types.StringType,
	}
}

func (m PostgresqlParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	postgresqlParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Postgresql = postgresqlParams
	return o, diag.Diagnostics{}
}

func (m PostgresqlParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Postgresql.IsNull()
}

func (m *PostgresqlParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Postgresql.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicPostgresqlParametersDto{
		Type:     sifflet.PublicPostgresqlParametersDtoTypePOSTGRESQL,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
	}
	err := parametersDto.FromPublicPostgresqlParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *PostgresqlParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Postgresql.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicPostgresqlParametersDto{
		Type:     sifflet.PublicPostgresqlParametersDtoTypePOSTGRESQL,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
	}
	err := parametersDto.FromPublicPostgresqlParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *PostgresqlParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicPostgresqlParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	m.Schema = types.StringPointerValue(paramsDto.Schema)
	return diag.Diagnostics{}
}

func (m PostgresqlParametersModel) RequiresCredential() bool {
	return true
}
