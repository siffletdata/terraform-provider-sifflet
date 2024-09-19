package source

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type MssqlParametersModel struct {
	Host     types.String `tfsdk:"host"`
	Database types.String `tfsdk:"database"`
	Port     types.Int32  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
	Ssl      types.Bool   `tfsdk:"ssl"`
}

func (m MssqlParametersModel) SchemaSourceType() string {
	return "mssql"
}

func (m MssqlParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Microsoft SQL Server hostname",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Microsoft SQL Server port number",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Schema name",
				Required:    true,
			},
			"ssl": schema.BoolAttribute{
				Description:        "Use TLS to connect to Microsoft SQL Server.",
				Optional:           true,
				Computed:           true,
				Default:            booldefault.StaticBool(true),
				DeprecationMessage: "Turning TLS off is for very specific use cases only and strongly discouraged. This option may be removed in the future.",
			},
		},
	}
}

func (m MssqlParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":     types.StringType,
		"database": types.StringType,
		"port":     types.Int32Type,
		"schema":   types.StringType,
		"ssl":      types.BoolType,
	}
}

func (m MssqlParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	mssqlParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Mssql = mssqlParams
	return o, diag.Diagnostics{}
}

func (m MssqlParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Mssql.IsNull()
}

func (m *MssqlParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Mssql.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicMssqlParametersDto{
		Type:     sifflet.PublicMssqlParametersDtoTypeMSSQL,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
		Ssl:      m.Ssl.ValueBoolPointer(),
	}
	err := parametersDto.FromPublicMssqlParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *MssqlParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Mssql.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicMssqlParametersDto{
		Type:     sifflet.PublicMssqlParametersDtoTypeMSSQL,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
		Ssl:      m.Ssl.ValueBoolPointer(),
	}
	err := parametersDto.FromPublicMssqlParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *MssqlParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicMssqlParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	m.Schema = types.StringPointerValue(paramsDto.Schema)
	m.Ssl = types.BoolPointerValue(paramsDto.Ssl)
	return diag.Diagnostics{}
}

func (m MssqlParametersModel) RequiresCredential() bool {
	return true
}
