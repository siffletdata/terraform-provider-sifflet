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

type AthenaParametersModel struct {
	Database         types.String `tfsdk:"database"`
	Datasource       types.String `tfsdk:"datasource"`
	Region           types.String `tfsdk:"region"`
	RoleArn          types.String `tfsdk:"role_arn"`
	S3OutputLocation types.String `tfsdk:"s3_output_location"`
	VpcUrl           types.String `tfsdk:"vpc_url"`
	Workgroup        types.String `tfsdk:"workgroup"`
}

func (m AthenaParametersModel) SchemaSourceType() string {
	return "athena"
}

func (m AthenaParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				Description: "Athena database name",
				Required:    true,
			},
			"datasource": schema.StringAttribute{
				Description: "Athena datasource name",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "AWS region in which the Athena database is located",
				Required:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "AWS IAM role ARN to use for Athena queries",
				Required:    true,
			},
			"s3_output_location": schema.StringAttribute{
				Description: "S3 location to store Athena query results",
				Required:    true,
			},
			"vpc_url": schema.StringAttribute{
				Description: "VPC URL for Athena queries",
				Optional:    true,
			},
			"workgroup": schema.StringAttribute{
				Description: "Athena workgroup name",
				Required:    true,
			},
		},
	}
}

func (m AthenaParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"database":           types.StringType,
		"datasource":         types.StringType,
		"region":             types.StringType,
		"role_arn":           types.StringType,
		"s3_output_location": types.StringType,
		"vpc_url":            types.StringType,
		"workgroup":          types.StringType,
	}
}

func (m AthenaParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	athenaParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Athena = athenaParams
	return o, diag.Diagnostics{}
}

func (m AthenaParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Athena.IsNull()
}

func (m *AthenaParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Athena.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicAthenaParametersDto{
		Type:             sifflet.PublicAthenaParametersDtoTypeATHENA,
		Database:         m.Database.ValueStringPointer(),
		Datasource:       m.Datasource.ValueStringPointer(),
		Region:           m.Region.ValueStringPointer(),
		RoleArn:          m.RoleArn.ValueStringPointer(),
		S3OutputLocation: m.S3OutputLocation.ValueStringPointer(),
		VpcUrl:           m.VpcUrl.ValueStringPointer(),
		Workgroup:        m.Workgroup.ValueStringPointer(),
	}
	err := parametersDto.FromPublicAthenaParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *AthenaParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicAthenaParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Database = types.StringPointerValue(paramsDto.Database)
	m.Datasource = types.StringPointerValue(paramsDto.Datasource)
	m.Region = types.StringPointerValue(paramsDto.Region)
	m.RoleArn = types.StringPointerValue(paramsDto.RoleArn)
	m.S3OutputLocation = types.StringPointerValue(paramsDto.S3OutputLocation)
	m.VpcUrl = types.StringPointerValue(paramsDto.VpcUrl)
	m.Workgroup = types.StringPointerValue(paramsDto.Workgroup)
	return diag.Diagnostics{}
}

func (m AthenaParametersModel) RequiresCredential() bool {
	return false
}
