package parameters_v2

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PostgresqlParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Database    types.String `tfsdk:"database"`
	Port        types.Int32  `tfsdk:"port"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m PostgresqlParametersModel) SchemaSourceType() string {
	return "postgresql"
}

func (m PostgresqlParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "PostgreSQL server host",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "PostgreSQL database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "PostgreSQL server port",
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
		},
	}
}

func (m PostgresqlParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"database":    types.StringType,
		"port":        types.Int32Type,
		"credentials": types.StringType,
		"schedule":    types.StringType,
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

func (m PostgresqlParametersModel) ToCreateDto(ctx context.Context, name string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	postgresqlInformation := sifflet.PostgresqlInformation{
		Host:     m.Host.ValueString(),
		Database: m.Database.ValueString(),
		Port:     m.Port.ValueInt32(),
	}

	postgresqlCreateDto := &sifflet.PublicCreatePostgresqlSourceV2Dto{
		Name:                  name,
		Type:                  sifflet.PublicCreatePostgresqlSourceV2DtoTypePOSTGRESQL,
		PostgresqlInformation: &postgresqlInformation,
		Credentials:           m.Credentials.ValueStringPointer(),
		Schedule:              m.Schedule.ValueStringPointer(),
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err := createSourceJsonBody.FromAny(postgresqlCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create PostgreSQL source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m PostgresqlParametersModel) ToUpdateDto(ctx context.Context, name string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	postgresqlInformation := sifflet.PostgresqlInformation{
		Host:     m.Host.ValueString(),
		Database: m.Database.ValueString(),
		Port:     m.Port.ValueInt32(),
	}

	postgresqlUpdateDto := &sifflet.PublicUpdatePostgresqlSourceV2Dto{
		Name:                  &name,
		Type:                  sifflet.PublicUpdatePostgresqlSourceV2DtoTypePOSTGRESQL,
		PostgresqlInformation: &postgresqlInformation,
		Credentials:           m.Credentials.ValueStringPointer(),
		Schedule:              m.Schedule.ValueStringPointer(),
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err := editSourceJsonBody.FromAny(postgresqlUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update PostgreSQL source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *PostgresqlParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	postgresqlDto := d.PublicGetPostgresqlSourceV2Dto
	if postgresqlDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read PostgreSQL source", "Source does not contain PostgreSQL params but was interpreted as a PostgreSQL source")}
	}

	m.Host = types.StringValue(postgresqlDto.PostgresqlInformation.Host)
	m.Database = types.StringValue(postgresqlDto.PostgresqlInformation.Database)
	m.Port = types.Int32Value(postgresqlDto.PostgresqlInformation.Port)
	m.Credentials = types.StringPointerValue(postgresqlDto.Credentials)
	m.Schedule = types.StringPointerValue(postgresqlDto.Schedule)
	return diag.Diagnostics{}
}
