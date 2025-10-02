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
				Description: "Your dbt project name (the 'name' value in your dbt_project.yml file)",
				Required:    true,
			},
			"target": schema.StringAttribute{
				Description: "Your dbt target name (the 'target' value in your profiles.yml file)",
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

func (m DbtParametersModel) ToCreateDto(ctx context.Context, name string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	dbtInformation := sifflet.DbtInformation{
		ProjectName: m.ProjectName.ValueString(),
		Target:      m.Target.ValueString(),
	}

	dbtCreateDto := &sifflet.PublicCreateDbtSourceV2Dto{
		Name:           name,
		Type:           sifflet.PublicCreateDbtSourceV2DtoTypeDBT,
		DbtInformation: &dbtInformation,
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err := createSourceJsonBody.FromAny(dbtCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create dbt source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m DbtParametersModel) ToUpdateDto(ctx context.Context, name string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	dbtInformation := sifflet.DbtInformation{
		ProjectName: m.ProjectName.ValueString(),
		Target:      m.Target.ValueString(),
	}

	dbtUpdateDto := &sifflet.PublicUpdateDbtSourceV2Dto{
		Name:           &name,
		Type:           sifflet.PublicUpdateDbtSourceV2DtoTypeDBT,
		DbtInformation: dbtInformation,
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err := editSourceJsonBody.FromAny(dbtUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update dbt source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *DbtParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	dbtDto := d.PublicGetDbtSourceV2Dto
	if dbtDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read dbt source", "Source does not contain dbt params but was interpreted as a dbt source")}
	}

	m.ProjectName = types.StringValue(dbtDto.DbtInformation.ProjectName)
	m.Target = types.StringValue(dbtDto.DbtInformation.Target)
	return diag.Diagnostics{}
}
