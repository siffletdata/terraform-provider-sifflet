package parameters_v2

import (
	"context"
	"encoding/json"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/source_v2/parameters_v2/scope"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OracleParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Database    types.String `tfsdk:"database"`
	Port        types.Int32  `tfsdk:"port"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
	Scope       types.Object `tfsdk:"scope"`
}

func (m OracleParametersModel) SchemaSourceType() string {
	return "oracle"
}

func (m OracleParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Oracle database host",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "Oracle database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Oracle database port",
				Required:    true,
			},
			"credentials": schema.StringAttribute{
				Description: "Name of the credentials used to connect to the source.",
				Required:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "Schedule for the source. Must be a valid cron expression. If empty, the source will only be refreshed when manually triggered.",
				Optional:    true,
			},
			"scope": schema.SingleNestedAttribute{
				Description: "Schemas to include or exclude. If not specified, all the schemas will be included (including future ones).",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Whether to include or exclude the specified schemas. One of INCLUSION or EXCLUSION.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INCLUSION", "EXCLUSION"),
						},
					},
					"schemas": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The schemas to either include or exclude.",
					},
				},
			},
		},
	}
}

func (m OracleParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"database":    types.StringType,
		"port":        types.Int32Type,
		"credentials": types.StringType,
		"schedule":    types.StringType,
		"scope":       scope.SchemasScopeTypeAttributes,
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

func (m OracleParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	oracleInformation := sifflet.OracleInformation{
		Host:     m.Host.ValueString(),
		Database: m.Database.ValueString(),
		Port:     m.Port.ValueInt32(),
	}

	scopeDto, diags := scope.ToPublicSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	oracleCreateDto := &sifflet.PublicCreateOracleSourceV2Dto{
		Name:              name,
		Timezone:          &timezone,
		Type:              sifflet.PublicCreateOracleSourceV2DtoTypeORACLE,
		OracleInformation: &oracleInformation,
		Credentials:       m.Credentials.ValueStringPointer(),
		Schedule:          m.Schedule.ValueStringPointer(),
		Scope:             scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(oracleCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Oracle source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m OracleParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	oracleInformation := sifflet.OracleInformation{
		Host:     m.Host.ValueString(),
		Database: m.Database.ValueString(),
		Port:     m.Port.ValueInt32(),
	}

	scopeDto, diags := scope.ToPublicSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	oracleUpdateDto := &sifflet.PublicUpdateOracleSourceV2Dto{
		Name:              &name,
		Timezone:          &timezone,
		Type:              sifflet.PublicUpdateOracleSourceV2DtoTypeORACLE,
		OracleInformation: oracleInformation,
		Credentials:       m.Credentials.ValueString(),
		Schedule:          m.Schedule.ValueStringPointer(),
		Scope:             scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(oracleUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Oracle source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *OracleParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	oracleDto := d.PublicGetOracleSourceV2Dto
	if oracleDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Oracle source", "Source does not contain Oracle params but was interpreted as a Oracle source")}
	}

	m.Host = types.StringValue(oracleDto.OracleInformation.Host)
	m.Database = types.StringValue(oracleDto.OracleInformation.Database)
	m.Port = types.Int32Value(oracleDto.OracleInformation.Port)
	m.Credentials = types.StringPointerValue(oracleDto.Credentials)
	m.Schedule = types.StringPointerValue(oracleDto.Schedule)
	scopeObject, diags := scope.FromPublicSchemasScopeDto(ctx, oracleDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}
