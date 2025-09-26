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

type FivetranParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m FivetranParametersModel) SchemaSourceType() string {
	return "fivetran"
}

func (m FivetranParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Your Fivetran environment URL",
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

func (m FivetranParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
	}
}

func (m FivetranParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	fivetranParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Fivetran = fivetranParams
	return o, diag.Diagnostics{}
}

func (m FivetranParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	fivetranInformation := sifflet.FivetranInformation{
		Host: m.Host.ValueString(),
	}

	fivetranCreateDto := &sifflet.PublicCreateFivetranSourceV2Dto{
		Name:                name,
		Timezone:            &timezone,
		Type:                sifflet.PublicCreateFivetranSourceV2DtoTypeFIVETRAN,
		FivetranInformation: &fivetranInformation,
		Credentials:         m.Credentials.ValueStringPointer(),
		Schedule:            m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(fivetranCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Fivetran source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m FivetranParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	fivetranInformation := sifflet.FivetranInformation{
		Host: m.Host.ValueString(),
	}

	fivetranUpdateDto := &sifflet.PublicUpdateFivetranSourceV2Dto{
		Name:                &name,
		Timezone:            &timezone,
		Type:                sifflet.PublicUpdateFivetranSourceV2DtoTypeFIVETRAN,
		FivetranInformation: fivetranInformation,
		Credentials:         m.Credentials.ValueString(),
		Schedule:            m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(fivetranUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Fivetran source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *FivetranParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	fivetranDto := d.PublicGetFivetranSourceV2Dto
	if fivetranDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Fivetran source", "Source does not contain Fivetran params but was interpreted as a Fivetran source")}
	}

	m.Host = types.StringValue(fivetranDto.FivetranInformation.Host)
	m.Credentials = types.StringPointerValue(fivetranDto.Credentials)
	m.Schedule = types.StringPointerValue(fivetranDto.Schedule)
	return diag.Diagnostics{}
}
