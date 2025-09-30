package parameters_v2

import (
	"context"
	"encoding/json"

	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AirflowParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Port        types.Int32  `tfsdk:"port"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m AirflowParametersModel) SchemaSourceType() string {
	return "airflow"
}

func (m AirflowParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Airflow server hostname",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Airflow server port",
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

func (m AirflowParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"port":        types.Int32Type,
		"credentials": types.StringType,
		"schedule":    types.StringType,
	}
}

func (m AirflowParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	airflowParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Airflow = airflowParams
	return o, diag.Diagnostics{}
}

func (m AirflowParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	airflowInformation := sifflet.AirflowInformation{
		Host: m.Host.ValueString(),
		Port: m.Port.ValueInt32(),
	}

	airflowCreateDto := sifflet.PublicCreateAirflowSourceV2Dto{
		Name:               name,
		Type:               sifflet.PublicCreateAirflowSourceV2DtoTypeAIRFLOW,
		Timezone:           &timezone,
		AirflowInformation: &airflowInformation,
		Credentials:        m.Credentials.ValueStringPointer(),
		Schedule:           m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(airflowCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Airflow source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m AirflowParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	airflowInformation := sifflet.AirflowInformation{
		Host: m.Host.ValueString(),
		Port: m.Port.ValueInt32(),
	}

	airflowUpdateDto := sifflet.PublicUpdateAirflowSourceV2Dto{
		Name:               &name,
		Type:               sifflet.PublicUpdateAirflowSourceV2DtoTypeAIRFLOW,
		Timezone:           &timezone,
		AirflowInformation: airflowInformation,
		Credentials:        m.Credentials.ValueString(),
		Schedule:           m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(airflowUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Airflow source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *AirflowParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	airflowDto := d.PublicGetAirflowSourceV2Dto
	if airflowDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Airflow source", "Source does not contain Airflow params but was interpreted as an Airflow source")}
	}

	m.Host = types.StringValue(airflowDto.AirflowInformation.Host)
	m.Port = types.Int32Value(airflowDto.AirflowInformation.Port)
	m.Credentials = types.StringPointerValue(airflowDto.Credentials)
	m.Schedule = types.StringPointerValue(airflowDto.Schedule)

	return diag.Diagnostics{}
}
