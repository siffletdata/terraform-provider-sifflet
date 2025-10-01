package parameters_v2

import (
	"context"
	"fmt"

	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MysqlParametersModel struct {
	Host            types.String `tfsdk:"host"`
	Database        types.String `tfsdk:"database"`
	Port            types.Int32  `tfsdk:"port"`
	MysqlTlsVersion types.String `tfsdk:"mysql_tls_version"`
	Credentials     types.String `tfsdk:"credentials"`
	Schedule        types.String `tfsdk:"schedule"`
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
				Description: "MySQL server port",
				Required:    true,
			},
			"mysql_tls_version": schema.StringAttribute{
				Description: "TLS version to use for MySQL connection. One of TLS_V_1_2 or TLS_V_1_3.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("TLS_V_1_2", "TLS_V_1_3"),
				},
			},
			"credentials": schema.StringAttribute{
				Description: "Name of the credentials used to connect to the source.",
				Required:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "Schedule for the source. Must be a valid cron expression. If empty, the source will only be refreshed when manually triggered.",
				Optional:    true,
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
		"credentials":       types.StringType,
		"schedule":          types.StringType,
	}
}

func parseMysqlTlsVersion(v string) (sifflet.MysqlInformationMysqlTlsVersion, error) {
	switch v {
	case "TLS_V_1_2":
		return sifflet.MysqlInformationMysqlTlsVersionTLSV12, nil
	case "TLS_V_1_3":
		return sifflet.MysqlInformationMysqlTlsVersionTLSV13, nil
	default:
		return "", fmt.Errorf("unsupported value for MySQL TLS version: %s", v)
	}
}

func mysqlTlsVersionToString(v sifflet.MysqlInformationMysqlTlsVersion) (string, error) {
	switch v {
	case sifflet.MysqlInformationMysqlTlsVersionTLSV12:
		return "TLS_V_1_2", nil
	case sifflet.MysqlInformationMysqlTlsVersionTLSV13:
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

func (m MysqlParametersModel) ToCreateDto(ctx context.Context, name string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	mysqlTlsVersion, err := parseMysqlTlsVersion(m.MysqlTlsVersion.ValueString())
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to create source", err.Error())}
	}
	mysqlInformation := sifflet.MysqlInformation{
		Database:        m.Database.ValueString(),
		Host:            m.Host.ValueString(),
		Port:            m.Port.ValueInt32(),
		MysqlTlsVersion: mysqlTlsVersion,
	}

	mysqlCreateDto := sifflet.PublicCreateMysqlSourceV2Dto{
		Name:             name,
		Type:             sifflet.PublicCreateMysqlSourceV2DtoTypeMYSQL,
		MysqlInformation: &mysqlInformation,
		Credentials:      m.Credentials.ValueStringPointer(),
		Schedule:         m.Schedule.ValueStringPointer(),
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err = createSourceJsonBody.FromAny(mysqlCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Mysql source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m MysqlParametersModel) ToUpdateDto(ctx context.Context, name string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	mysqlTlsVersion, err := parseMysqlTlsVersion(m.MysqlTlsVersion.ValueString())
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, diag.Diagnostics{diag.NewErrorDiagnostic("Unable to create source", err.Error())}
	}
	mysqlInformation := sifflet.MysqlInformation{
		Database:        m.Database.ValueString(),
		Host:            m.Host.ValueString(),
		Port:            m.Port.ValueInt32(),
		MysqlTlsVersion: mysqlTlsVersion,
	}

	mysqlUpdateDto := sifflet.PublicUpdateMysqlSourceV2Dto{
		Name:             &name,
		Type:             sifflet.PublicUpdateMysqlSourceV2DtoTypeMYSQL,
		MysqlInformation: mysqlInformation,
		Credentials:      m.Credentials.ValueString(),
		Schedule:         m.Schedule.ValueStringPointer(),
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err = editSourceJsonBody.FromAny(mysqlUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Mysql source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *MysqlParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	mysqlDto := d.PublicGetMysqlSourceV2Dto
	if mysqlDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read source", "Source does not contain Mysql params but was interpreted as a Mysql source")}
	}

	m.Host = types.StringValue(mysqlDto.MysqlInformation.Host)
	m.Database = types.StringValue(mysqlDto.MysqlInformation.Database)
	m.Port = types.Int32Value(mysqlDto.MysqlInformation.Port)
	mysqlTlsVersion, err := mysqlTlsVersionToString(mysqlDto.MysqlInformation.MysqlTlsVersion)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Mysql source", err.Error())}
	}
	m.MysqlTlsVersion = types.StringValue(mysqlTlsVersion)
	m.Credentials = types.StringPointerValue(mysqlDto.Credentials)
	m.Schedule = types.StringPointerValue(mysqlDto.Schedule)
	return diag.Diagnostics{}
}
