package parameters

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type FivetranParametersModel struct {
	Host types.String `tfsdk:"host"`
}

func (m FivetranParametersModel) SchemaSourceType() string {
	return "fivetran"
}

func (m FivetranParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Fivetran host. Defaults to https://api.fivetran.com.",
				Optional:    true,
				Default:     stringdefault.StaticString("https://api.fivetran.com"),
				Computed:    true,
			},
		},
	}
}

func (m FivetranParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host": types.StringType,
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

func (m FivetranParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Fivetran.IsNull()
}

func (m *FivetranParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Fivetran.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicFivetranParametersDto{
		Host: m.Host.ValueStringPointer(),
		Type: sifflet.PublicFivetranParametersDtoTypeFIVETRAN,
	}
	err := parametersDto.FromPublicFivetranParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *FivetranParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Fivetran.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicFivetranParametersDto{
		Host: m.Host.ValueStringPointer(),
		Type: sifflet.PublicFivetranParametersDtoTypeFIVETRAN,
	}
	err := parametersDto.FromPublicFivetranParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *FivetranParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicFivetranParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	return diag.Diagnostics{}
}

func (m FivetranParametersModel) RequiresCredential() bool {
	return true
}
