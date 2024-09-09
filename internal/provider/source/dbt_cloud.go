package source

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DbtCloudParametersModel struct {
	AccountID       types.String `tfsdk:"account_id"`
	BaseUrl         types.String `tfsdk:"base_url"`
	JobDefinitionID types.String `tfsdk:"job_definition_id"`
	ProjectID       types.String `tfsdk:"project_id"`
}

func (m DbtCloudParametersModel) SchemaSourceType() string {
	return "dbt_cloud"
}

func (m DbtCloudParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "dbt Cloud account ID",
				Required:    true,
			},
			"base_url": schema.StringAttribute{
				Description: "dbt Cloud base URL",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "dbt Cloud project ID",
				Required:    true,
			},
			"job_definition_id": schema.StringAttribute{
				Description: "dbt Cloud job definition ID",
				Optional:    true,
			},
		},
	}
}

func (m DbtCloudParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_id":        types.StringType,
		"base_url":          types.StringType,
		"project_id":        types.StringType,
		"job_definition_id": types.StringType,
	}
}

func (m DbtCloudParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	dbtcloudParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.DbtCloud = dbtcloudParams
	return o, diag.Diagnostics{}
}

func (m DbtCloudParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.DbtCloud.IsNull()
}

func (m *DbtCloudParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.DbtCloud.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicDbtCloudParametersDto{
		Type:            sifflet.PublicDbtCloudParametersDtoTypeDBTCLOUD,
		AccountId:       m.AccountID.ValueStringPointer(),
		BaseUrl:         m.BaseUrl.ValueStringPointer(),
		ProjectId:       m.ProjectID.ValueStringPointer(),
		JobDefinitionId: m.JobDefinitionID.ValueStringPointer(),
	}
	err := parametersDto.FromPublicDbtCloudParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *DbtCloudParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicDbtCloudParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.AccountID = types.StringPointerValue(paramsDto.AccountId)
	m.BaseUrl = types.StringPointerValue(paramsDto.BaseUrl)
	m.ProjectID = types.StringPointerValue(paramsDto.ProjectId)
	m.JobDefinitionID = types.StringPointerValue(paramsDto.JobDefinitionId)
	return diag.Diagnostics{}
}

func (m DbtCloudParametersModel) RequiresCredential() bool {
	return true
}
