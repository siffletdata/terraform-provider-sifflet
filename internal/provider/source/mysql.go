package source

import (
	"context"
	"fmt"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type MysqlParametersModel struct {
	Host            types.String `tfsdk:"host"`
	Database        types.String `tfsdk:"database"`
	Port            types.Int32  `tfsdk:"port"`
	MysqlTlsVersion types.String `tfsdk:"mysql_tls_version"`
}

func (m MysqlParametersModel) SchemaSourceType() string {
	return "mysql"
}

func (m MysqlParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "MySQL server hostname",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "MySQL port number",
				Required:    true,
			},
			"mysql_tls_version": schema.StringAttribute{
				Description: "TLS version to use for MySQL connection. One of TLS_V_1_2 or TLS_V_1_3.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("TLS_V_1_2", "TLS_V_1_3"),
				},
			},
		},
	}
}

func (m MysqlParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":              types.StringType,
		"database":          types.StringType,
		"port":              types.Int32Type,
		"mysql_tls_version": types.StringType,
	}
}

func parseMysqlTlsVersion(v string) (sifflet.PublicMysqlParametersDtoMysqlTlsVersion, error) {
	switch v {
	case "TLS_V_1_2":
		return sifflet.TLSV12, nil
	case "TLS_V_1_3":
		return sifflet.TLSV13, nil
	default:
		return "", fmt.Errorf("unsupported value for MySQL TLS version: %s", v)
	}
}

func mysqlTlsVersionToString(v sifflet.PublicMysqlParametersDtoMysqlTlsVersion) (string, error) {
	switch v {
	case sifflet.TLSV12:
		return "TLS_V_1_2", nil
	case sifflet.TLSV13:
		return "TLS_V_1_3", nil
	default:
		return "", fmt.Errorf("unsupported value for MySQL TLS version: %s", v)
	}
}

func (m MysqlParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	mysqlParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Mysql = mysqlParams
	return o, diag.Diagnostics{}
}

func (m MysqlParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Mysql.IsNull()
}

func (m *MysqlParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Mysql.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	mysqlTlsVersion, err := parseMysqlTlsVersion(m.MysqlTlsVersion.ValueString())
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	dto := sifflet.PublicMysqlParametersDto{
		Type:            sifflet.PublicMysqlParametersDtoTypeMYSQL,
		Host:            m.Host.ValueStringPointer(),
		Database:        m.Database.ValueStringPointer(),
		Port:            m.Port.ValueInt32Pointer(),
		MysqlTlsVersion: &mysqlTlsVersion,
	}
	err = parametersDto.FromPublicMysqlParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *MysqlParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Mysql.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	mysqlTlsVersion, err := parseMysqlTlsVersion(m.MysqlTlsVersion.ValueString())
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	dto := sifflet.PublicMysqlParametersDto{
		Type:            sifflet.PublicMysqlParametersDtoTypeMYSQL,
		Host:            m.Host.ValueStringPointer(),
		Database:        m.Database.ValueStringPointer(),
		Port:            m.Port.ValueInt32Pointer(),
		MysqlTlsVersion: &mysqlTlsVersion,
	}
	err = parametersDto.FromPublicMysqlParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *MysqlParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicMysqlParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	mysqlTlsVersion, err := mysqlTlsVersionToString(*paramsDto.MysqlTlsVersion)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to create source", err.Error())}
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Port = types.Int32PointerValue(paramsDto.Port)
	m.MysqlTlsVersion = types.StringValue(mysqlTlsVersion)
	return diag.Diagnostics{}
}

func (m MysqlParametersModel) RequiresCredential() bool {
	return true
}
