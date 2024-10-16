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

type DbtParametersModel struct {
	ProjectName types.String `tfsdk:"project_name"`
	Target      types.String `tfsdk:"target"`
}

func (m DbtParametersModel) SchemaSourceType() string {
	return "dbt"
}

func (m DbtParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Description: "dbt project name",
				Required:    true,
			},
			"target": schema.StringAttribute{
				Description: "dbt target name (the 'target' value in the profiles.yml file)",
				Required:    true,
			},
		},
	}
}

func (m DbtParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"project_name": types.StringType,
		"target":       types.StringType,
	}
}

func (m DbtParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	dbtParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Dbt = dbtParams
	return o, diag.Diagnostics{}
}

func (m DbtParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Dbt.IsNull()
}

func (m *DbtParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Dbt.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicDbtParametersDto{
		Type:        sifflet.PublicDbtParametersDtoTypeDBT,
		ProjectName: m.ProjectName.ValueStringPointer(),
		Target:      m.Target.ValueStringPointer(),
	}
	err := parametersDto.FromPublicDbtParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *DbtParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.Dbt.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicDbtParametersDto{
		Type:        sifflet.PublicDbtParametersDtoTypeDBT,
		ProjectName: m.ProjectName.ValueStringPointer(),
		Target:      m.Target.ValueStringPointer(),
	}
	err := parametersDto.FromPublicDbtParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *DbtParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicDbtParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.ProjectName = types.StringPointerValue(paramsDto.ProjectName)
	m.Target = types.StringPointerValue(paramsDto.Target)
	return diag.Diagnostics{}
}

func (m DbtParametersModel) RequiresCredential() bool {
	return false
}
