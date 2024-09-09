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

type QuickSightParametersModel struct {
	AccountID types.String `tfsdk:"account_id"`
	AwsRegion types.String `tfsdk:"aws_region"`
	RoleArn   types.String `tfsdk:"role_arn"`
}

func (m QuickSightParametersModel) SchemaSourceType() string {
	return "quicksight"
}

func (m QuickSightParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "AWS account ID",
				Required:    true,
			},
			"aws_region": schema.StringAttribute{
				Description: "AWS region",
				Required:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "AWS IAM role ARN used to access QuickSight (see Sifflet documentation for details)",
				Required:    true,
			},
		},
	}
}

func (m QuickSightParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_id": types.StringType,
		"aws_region": types.StringType,
		"role_arn":   types.StringType,
	}
}

func (m QuickSightParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	quicksightParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.QuickSight = quicksightParams
	return o, diag.Diagnostics{}
}

func (m QuickSightParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.QuickSight.IsNull()
}

func (m *QuickSightParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.QuickSight.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicQuicksightParametersDto{
		Type:      sifflet.PublicQuicksightParametersDtoTypeQUICKSIGHT,
		AccountId: m.AccountID.ValueStringPointer(),
		AwsRegion: m.AwsRegion.ValueStringPointer(),
		RoleArn:   m.RoleArn.ValueStringPointer(),
	}
	err := parametersDto.FromPublicQuicksightParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *QuickSightParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicQuicksightParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.AccountID = types.StringPointerValue(paramsDto.AccountId)
	m.AwsRegion = types.StringPointerValue(paramsDto.AwsRegion)
	m.RoleArn = types.StringPointerValue(paramsDto.RoleArn)
	return diag.Diagnostics{}
}

func (m QuickSightParametersModel) RequiresCredential() bool {
	return false
}
