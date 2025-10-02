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

type PowerBiParametersModel struct {
	ClientId    types.String `tfsdk:"client_id"`
	TenantId    types.String `tfsdk:"tenant_id"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m PowerBiParametersModel) SchemaSourceType() string {
	return "power_bi"
}

func (m PowerBiParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "Your Azure AD client ID",
				Required:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Your Azure AD tenant ID",
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

func (m PowerBiParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":   types.StringType,
		"tenant_id":   types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
	}
}

func (m PowerBiParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	powerBiParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.PowerBi = powerBiParams
	return o, diag.Diagnostics{}
}

func (m PowerBiParametersModel) ToCreateDto(ctx context.Context, name string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	powerBiInformation := sifflet.PowerBiInformation{
		ClientId: m.ClientId.ValueString(),
		TenantId: m.TenantId.ValueString(),
	}

	powerBiCreateDto := &sifflet.PublicCreatePowerBiSourceV2Dto{
		Name:               name,
		Type:               sifflet.PublicCreatePowerBiSourceV2DtoTypePOWERBI,
		PowerBiInformation: &powerBiInformation,
		Credentials:        m.Credentials.ValueStringPointer(),
		Schedule:           m.Schedule.ValueStringPointer(),
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err := createSourceJsonBody.FromAny(powerBiCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Power BI source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m PowerBiParametersModel) ToUpdateDto(ctx context.Context, name string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	powerBiInformation := sifflet.PowerBiInformation{
		ClientId: m.ClientId.ValueString(),
		TenantId: m.TenantId.ValueString(),
	}

	powerBiUpdateDto := &sifflet.PublicUpdatePowerBiSourceV2Dto{
		Name:               &name,
		Type:               sifflet.PublicUpdatePowerBiSourceV2DtoTypePOWERBI,
		PowerBiInformation: powerBiInformation,
		Credentials:        m.Credentials.ValueString(),
		Schedule:           m.Schedule.ValueStringPointer(),
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err := editSourceJsonBody.FromAny(powerBiUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Power BI source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *PowerBiParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	powerBiDto := d.PublicGetPowerBiSourceV2Dto
	if powerBiDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Power BI source", "Source does not contain Power BI params but was interpreted as a Power BI source")}
	}

	m.ClientId = types.StringValue(powerBiDto.PowerBiInformation.ClientId)
	m.TenantId = types.StringValue(powerBiDto.PowerBiInformation.TenantId)
	m.Credentials = types.StringPointerValue(powerBiDto.Credentials)
	m.Schedule = types.StringPointerValue(powerBiDto.Schedule)
	return diag.Diagnostics{}
}
