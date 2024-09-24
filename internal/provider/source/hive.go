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

type HiveParametersModel struct {
	AtlasBaseUrl   types.String `tfsdk:"atlas_base_url"`
	AtlasPrincipal types.String `tfsdk:"atlas_principal"`
	Database       types.String `tfsdk:"database"`
	JdbcUrl        types.String `tfsdk:"jdbc_url"`
	Krb5Conf       types.String `tfsdk:"krb5_conf"`
	Principal      types.String `tfsdk:"principal"`
}

func (m HiveParametersModel) SchemaSourceType() string {
	return "hive"
}

func (m HiveParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"atlas_base_url": schema.StringAttribute{
				Description: "Atlas server base URL",
				Optional:    true,
			},
			"atlas_principal": schema.StringAttribute{
				Description: "Atlas server principal",
				Optional:    true,
			},
			"database": schema.StringAttribute{
				Description: "Hive database name",
				Required:    true,
			},
			"jdbc_url": schema.StringAttribute{
				Description: "Hive server JDBC URL",
				Required:    true,
			},
			"krb5_conf": schema.StringAttribute{
				Description: "Kerberos configuration file (krb5.conf)",
				Required:    true,
			},
			"principal": schema.StringAttribute{
				Description: "Hive server principal",
				Required:    true,
			},
		},
	}
}

func (m HiveParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"atlas_base_url":  types.StringType,
		"atlas_principal": types.StringType,
		"database":        types.StringType,
		"jdbc_url":        types.StringType,
		"krb5_conf":       types.StringType,
		"principal":       types.StringType,
	}
}

func (m HiveParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	hiveParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Hive = hiveParams
	return o, diag.Diagnostics{}
}

func (m HiveParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Hive.IsNull()
}

func (m *HiveParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Hive.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}

	dto := sifflet.PublicHiveParametersDto{
		AtlasBaseUrl:   m.AtlasBaseUrl.ValueStringPointer(),
		AtlasPrincipal: m.AtlasPrincipal.ValueStringPointer(),
		Database:       m.Database.ValueStringPointer(),
		JdbcUrl:        m.JdbcUrl.ValueStringPointer(),
		Krb5Conf:       m.Krb5Conf.ValueStringPointer(),
		Principal:      m.Principal.ValueStringPointer(),
		Type:           sifflet.PublicHiveParametersDtoTypeHIVE,
	}
	err := parametersDto.FromPublicHiveParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *HiveParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Hive.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}

	dto := sifflet.PublicHiveParametersDto{
		AtlasBaseUrl:   m.AtlasBaseUrl.ValueStringPointer(),
		AtlasPrincipal: m.AtlasPrincipal.ValueStringPointer(),
		Database:       m.Database.ValueStringPointer(),
		JdbcUrl:        m.JdbcUrl.ValueStringPointer(),
		Krb5Conf:       m.Krb5Conf.ValueStringPointer(),
		Principal:      m.Principal.ValueStringPointer(),
		Type:           sifflet.PublicHiveParametersDtoTypeHIVE,
	}
	err := parametersDto.FromPublicHiveParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *HiveParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicHiveParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.AtlasBaseUrl = types.StringPointerValue(paramsDto.AtlasBaseUrl)
	m.AtlasPrincipal = types.StringPointerValue(paramsDto.AtlasPrincipal)
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.JdbcUrl = types.StringPointerValue(paramsDto.JdbcUrl)
	m.Krb5Conf = types.StringPointerValue(paramsDto.Krb5Conf)
	m.Principal = types.StringPointerValue(paramsDto.Principal)
	return diag.Diagnostics{}
}

func (m HiveParametersModel) RequiresCredential() bool {
	return true
}
