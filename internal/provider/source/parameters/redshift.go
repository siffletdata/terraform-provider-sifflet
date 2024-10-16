package parameters

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

type RedshiftParametersModel struct {
	Host     types.String `tfsdk:"host"`
	Database types.String `tfsdk:"database"`
	Port     types.Int32  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
	Ssl      types.Bool   `tfsdk:"ssl"`
}

func (m RedshiftParametersModel) SchemaSourceType() string {
	return "redshift"
}

func (m RedshiftParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Redshift server hostname",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Redshift server port number",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Schema name",
				Required:    true,
			},
			"ssl": schema.BoolAttribute{
				Description: "Use TLS to connect to Redshift. It's strongly recommended to keep this option enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (m RedshiftParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":     types.StringType,
		"database": types.StringType,
		"port":     types.Int32Type,
		"schema":   types.StringType,
		"ssl":      types.BoolType,
	}
}

func (m RedshiftParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	redshiftParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Redshift = redshiftParams
	return o, diag.Diagnostics{}
}

func (m RedshiftParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Redshift.IsNull()
}

func (m *RedshiftParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Redshift.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicRedshiftParametersDto{
		Type:     sifflet.PublicRedshiftParametersDtoTypeREDSHIFT,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
		Ssl:      m.Ssl.ValueBoolPointer(),
	}
	err := parametersDto.FromPublicRedshiftParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *RedshiftParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Redshift.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicRedshiftParametersDto{
		Type:     sifflet.PublicRedshiftParametersDtoTypeREDSHIFT,
		Host:     m.Host.ValueStringPointer(),
		Database: m.Database.ValueStringPointer(),
		Port:     m.Port.ValueInt32Pointer(),
		Schema:   m.Schema.ValueStringPointer(),
		Ssl:      m.Ssl.ValueBoolPointer(),
	}
	err := parametersDto.FromPublicRedshiftParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *RedshiftParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicRedshiftParametersDto()
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

func (m RedshiftParametersModel) RequiresCredential() bool {
	return true
}
