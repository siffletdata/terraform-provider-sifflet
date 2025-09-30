package parameters_v2

import (
	"context"
	"encoding/json"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MssqlParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Database    types.String `tfsdk:"database"`
	Port        types.Int32  `tfsdk:"port"`
	Ssl         types.Bool   `tfsdk:"ssl"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m MssqlParametersModel) SchemaSourceType() string {
	return "mssql"
}

func (m MssqlParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "MSSQL server hostname",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "MSSQL database name",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "MSSQL server port",
				Required:    true,
			},
			"credentials": schema.StringAttribute{
				Description: "Name of the credentials used to connect to the source.",
				Required:    true,
			},
			"ssl": schema.BoolAttribute{
				Description:        "Use TLS to connect to Microsoft SQL Server.",
				Optional:           true,
				Computed:           true,
				Default:            booldefault.StaticBool(true),
				DeprecationMessage: "Turning TLS off is for very specific use cases only and strongly discouraged. This option may be removed in the future.",
			},
			"schedule": schema.StringAttribute{
				Description: "Schedule for the source. Must be a valid cron expression. If empty, the source will only be refreshed when manually triggered.",
				Optional:    true,
			},
		},
	}
}

func (m MssqlParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"database":    types.StringType,
		"port":        types.Int32Type,
		"credentials": types.StringType,
		"ssl":         types.BoolType,
		"schedule":    types.StringType,
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

func (m MssqlParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	mssqlInformation := sifflet.MssqlInformation{
		Host:     m.Host.ValueString(),
		Database: m.Database.ValueString(),
		Port:     m.Port.ValueInt32(),
		Ssl:      m.Ssl.ValueBool(),
	}

	mssqlCreateDto := &sifflet.PublicCreateMssqlSourceV2Dto{
		Name:             name,
		Timezone:         &timezone,
		Type:             sifflet.PublicCreateMssqlSourceV2DtoTypeMSSQL,
		MssqlInformation: &mssqlInformation,
		Credentials:      m.Credentials.ValueStringPointer(),
		Schedule:         m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(mssqlCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create MSSQL source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m MssqlParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	mssqlInformation := sifflet.MssqlInformation{
		Host:     m.Host.ValueString(),
		Database: m.Database.ValueString(),
		Port:     m.Port.ValueInt32(),
		Ssl:      m.Ssl.ValueBool(),
	}

	mssqlUpdateDto := &sifflet.PublicUpdateMssqlSourceV2Dto{
		Name:             &name,
		Timezone:         &timezone,
		Type:             sifflet.PublicUpdateMssqlSourceV2DtoTypeMSSQL,
		MssqlInformation: mssqlInformation,
		Credentials:      m.Credentials.ValueString(),
		Schedule:         m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(mssqlUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update MSSQL source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *MssqlParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	mssqlDto := d.PublicGetMssqlSourceV2Dto
	if mssqlDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read MSSQL source", "Source does not contain MSSQL params but was interpreted as a MSSQL source")}
	}

	m.Host = types.StringValue(mssqlDto.MssqlInformation.Host)
	m.Database = types.StringValue(mssqlDto.MssqlInformation.Database)
	m.Port = types.Int32Value(mssqlDto.MssqlInformation.Port)
	m.Ssl = types.BoolValue(mssqlDto.MssqlInformation.Ssl)
	m.Credentials = types.StringPointerValue(mssqlDto.Credentials)
	m.Schedule = types.StringPointerValue(mssqlDto.Schedule)
	return diag.Diagnostics{}
}
