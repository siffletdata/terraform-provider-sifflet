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

type TableauParametersModel struct {
	Host types.String `tfsdk:"host"`
	Site types.String `tfsdk:"site"`
}

func (m TableauParametersModel) SchemaSourceType() string {
	return "tableau"
}

func (m TableauParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Tableau Server hostname",
				Required:    true,
			},
			"site": schema.StringAttribute{
				Description: "Tableau Server site. If your Tableau environment is using the Default site, omit this field.",
				Optional:    true,
			},
		},
	}
}

func (m TableauParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host": types.StringType,
		"site": types.StringType,
	}
}

func (m TableauParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	tableauParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Tableau = tableauParams
	return o, diag.Diagnostics{}
}

func (m TableauParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Tableau.IsNull()
}

func (m *TableauParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Tableau.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicTableauParametersDto{
		Type: sifflet.PublicTableauParametersDtoTypeTABLEAU,
		Host: m.Host.ValueStringPointer(),
		Site: m.Site.ValueStringPointer(),
	}
	err := parametersDto.FromPublicTableauParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *TableauParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Tableau.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicTableauParametersDto{
		Type: sifflet.PublicTableauParametersDtoTypeTABLEAU,
		Host: m.Host.ValueStringPointer(),
		Site: m.Site.ValueStringPointer(),
	}
	err := parametersDto.FromPublicTableauParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *TableauParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicTableauParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)
	m.Site = types.StringPointerValue(paramsDto.Site)
	return diag.Diagnostics{}
}

func (m TableauParametersModel) RequiresCredential() bool {
	return true
}
