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

type RedshiftParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Port        types.Int32  `tfsdk:"port"`
	Ssl         types.Bool   `tfsdk:"ssl"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m RedshiftParametersModel) SchemaSourceType() string {
	return "redshift"
}

func (m RedshiftParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Redshift cluster hostname",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Redshift cluster port",
				Required:    true,
			},
			"ssl": schema.BoolAttribute{
				Description: "Whether to use SSL to connect to your Redshift server",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
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

func (m RedshiftParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"port":        types.Int32Type,
		"ssl":         types.BoolType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
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

func (m RedshiftParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	redshiftInformation := sifflet.RedshiftInformation{
		Host: m.Host.ValueString(),
		Port: m.Port.ValueInt32(),
		Ssl:  m.Ssl.ValueBool(),
	}

	redshiftCreateDto := &sifflet.PublicCreateRedshiftSourceV2Dto{
		Name:                name,
		Timezone:            &timezone,
		Type:                sifflet.PublicCreateRedshiftSourceV2DtoTypeREDSHIFT,
		RedshiftInformation: &redshiftInformation,
		Credentials:         m.Credentials.ValueStringPointer(),
		Schedule:            m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(redshiftCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Redshift source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m RedshiftParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	redshiftInformation := sifflet.RedshiftInformation{
		Host: m.Host.ValueString(),
		Port: m.Port.ValueInt32(),
		Ssl:  m.Ssl.ValueBool(),
	}

	redshiftUpdateDto := &sifflet.PublicUpdateRedshiftSourceV2Dto{
		Name:                &name,
		Timezone:            &timezone,
		Type:                sifflet.PublicUpdateRedshiftSourceV2DtoTypeREDSHIFT,
		RedshiftInformation: redshiftInformation,
		Credentials:         m.Credentials.ValueString(),
		Schedule:            m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(redshiftUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Redshift source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *RedshiftParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	redshiftDto := d.PublicGetRedshiftSourceV2Dto
	if redshiftDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Redshift source", "Source does not contain Redshift params but was interpreted as a Redshift source")}
	}

	m.Host = types.StringValue(redshiftDto.RedshiftInformation.Host)
	m.Port = types.Int32Value(redshiftDto.RedshiftInformation.Port)
	m.Ssl = types.BoolValue(redshiftDto.RedshiftInformation.Ssl)
	m.Credentials = types.StringPointerValue(redshiftDto.Credentials)
	m.Schedule = types.StringPointerValue(redshiftDto.Schedule)
	return diag.Diagnostics{}
}
