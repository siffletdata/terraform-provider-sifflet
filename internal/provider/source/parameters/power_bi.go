package parameters

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type PowerBiParametersModel struct {
	ClientID    types.String `tfsdk:"client_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}

func (m PowerBiParametersModel) SchemaSourceType() string {
	return "power_bi"
}

func (m PowerBiParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "Azure AD client ID",
				Required:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Azure AD tenant ID",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Power BI workspace ID",
				Required:    true,
			},
		},
	}
}

func (m PowerBiParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":    types.StringType,
		"tenant_id":    types.StringType,
		"workspace_id": types.StringType,
	}
}

func (m PowerBiParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	powerbiParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.PowerBi = powerbiParams
	return o, diag.Diagnostics{}
}

func (m PowerBiParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.PowerBi.IsNull()
}

func (m *PowerBiParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.PowerBi.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicPowerBiParametersDto{
		Type:        sifflet.PublicPowerBiParametersDtoTypePOWERBI,
		ClientId:    m.ClientID.ValueStringPointer(),
		TenantId:    m.TenantID.ValueStringPointer(),
		WorkspaceId: m.WorkspaceID.ValueStringPointer(),
	}
	err := parametersDto.FromPublicPowerBiParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *PowerBiParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.PowerBi.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicPowerBiParametersDto{
		Type:        sifflet.PublicPowerBiParametersDtoTypePOWERBI,
		ClientId:    m.ClientID.ValueStringPointer(),
		TenantId:    m.TenantID.ValueStringPointer(),
		WorkspaceId: m.WorkspaceID.ValueStringPointer(),
	}
	err := parametersDto.FromPublicPowerBiParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *PowerBiParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicPowerBiParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.ClientID = types.StringPointerValue(paramsDto.ClientId)
	m.TenantID = types.StringPointerValue(paramsDto.TenantId)
	m.WorkspaceID = types.StringPointerValue(paramsDto.WorkspaceId)
	return diag.Diagnostics{}
}

func (m PowerBiParametersModel) RequiresCredential() bool {
	return true
}
