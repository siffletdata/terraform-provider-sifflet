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

type SnowflakeParametersModel struct {
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	Database          types.String `tfsdk:"database"`
	Schema            types.String `tfsdk:"schema"`
	Warehouse         types.String `tfsdk:"warehouse"`
}

func (m SnowflakeParametersModel) SchemaSourceType() string {
	return "snowflake"
}

func (m SnowflakeParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"account_identifier": schema.StringAttribute{
				Description: "Snowflake account identifier",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "Schema name",
				Required:    true,
			},
			"warehouse": schema.StringAttribute{
				Description: "Warehouse name, used by Sifflet to run queries",
				Required:    true,
			},
		},
	}
}

func (m SnowflakeParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_identifier": types.StringType,
		"database":           types.StringType,
		"schema":             types.StringType,
		"warehouse":          types.StringType,
	}
}

func (m SnowflakeParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	snowflakeParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Snowflake = snowflakeParams
	return o, diag.Diagnostics{}
}

func (m SnowflakeParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Snowflake.IsNull()
}

func (m *SnowflakeParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Snowflake.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicSnowflakeParametersDto{
		Type:              sifflet.PublicSnowflakeParametersDtoTypeSNOWFLAKE,
		AccountIdentifier: m.AccountIdentifier.ValueStringPointer(),
		Database:          m.Database.ValueStringPointer(),
		Schema:            m.Schema.ValueStringPointer(),
		Warehouse:         m.Warehouse.ValueStringPointer(),
	}
	err := parametersDto.FromPublicSnowflakeParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *SnowflakeParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicSnowflakeParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.AccountIdentifier = types.StringPointerValue(paramsDto.AccountIdentifier)
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Schema = types.StringPointerValue(paramsDto.Schema)
	m.Warehouse = types.StringPointerValue(paramsDto.Warehouse)
	return diag.Diagnostics{}
}

func (m SnowflakeParametersModel) RequiresCredential() bool {
	return true
}
