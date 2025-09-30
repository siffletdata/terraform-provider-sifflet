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

type AthenaParametersModel struct {
	Datasource       types.String `tfsdk:"datasource"`
	Region           types.String `tfsdk:"region"`
	RoleArn          types.String `tfsdk:"role_arn"`
	S3OutputLocation types.String `tfsdk:"s3_output_location"`
	VpcUrl           types.String `tfsdk:"vpc_url"`
	Workgroup        types.String `tfsdk:"workgroup"`
	Schedule         types.String `tfsdk:"schedule"`
}

func (m AthenaParametersModel) SchemaSourceType() string {
	return "athena"
}

func (m AthenaParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"datasource": schema.StringAttribute{
				Description: "Athena datasource name",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "AWS region in which the Athena instance is located",
				Required:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "AWS IAM role ARN to use for Athena queries",
				Required:    true,
			},
			"s3_output_location": schema.StringAttribute{
				Description: "The S3 location where Athena query results are stored",
				Required:    true,
			},
			"vpc_url": schema.StringAttribute{
				Description: "VPC URL for Athena connection",
				Optional:    true,
			},
			"workgroup": schema.StringAttribute{
				Description: "Athena workgroup name",
				Required:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "Schedule for the source. Must be a valid cron expression. If empty, the source will only be refreshed when manually triggered.",
				Optional:    true,
			},
		},
	}
}

func (m AthenaParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"datasource":         types.StringType,
		"region":             types.StringType,
		"role_arn":           types.StringType,
		"s3_output_location": types.StringType,
		"vpc_url":            types.StringType,
		"workgroup":          types.StringType,
		"schedule":           types.StringType,
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

func (m AthenaParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	athenaInformation := sifflet.AthenaInformation{
		Datasource:       m.Datasource.ValueString(),
		Region:           m.Region.ValueString(),
		RoleArn:          m.RoleArn.ValueString(),
		S3OutputLocation: m.S3OutputLocation.ValueString(),
		VpcUrl:           m.VpcUrl.ValueStringPointer(),
		Workgroup:        m.Workgroup.ValueString(),
	}

	athenaCreateDto := &sifflet.PublicCreateAthenaSourceV2Dto{
		Name:              name,
		Timezone:          &timezone,
		Type:              sifflet.PublicCreateAthenaSourceV2DtoTypeATHENA,
		AthenaInformation: &athenaInformation,
		Schedule:          m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(athenaCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Athena source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m AthenaParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	athenaInformation := sifflet.AthenaInformation{
		Datasource:       m.Datasource.ValueString(),
		Region:           m.Region.ValueString(),
		RoleArn:          m.RoleArn.ValueString(),
		S3OutputLocation: m.S3OutputLocation.ValueString(),
		VpcUrl:           m.VpcUrl.ValueStringPointer(),
		Workgroup:        m.Workgroup.ValueString(),
	}

	athenaUpdateDto := &sifflet.PublicUpdateAthenaSourceV2Dto{
		Name:              &name,
		Timezone:          &timezone,
		Type:              sifflet.PublicUpdateAthenaSourceV2DtoTypeATHENA,
		AthenaInformation: athenaInformation,
		Schedule:          m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(athenaUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Athena source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *AthenaParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	athenaDto := d.PublicGetAthenaSourceV2Dto
	if athenaDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Athena source", "Source does not contain Athena params but was interpreted as a Athena source")}
	}

	m.Datasource = types.StringValue(athenaDto.AthenaInformation.Datasource)
	m.Region = types.StringValue(athenaDto.AthenaInformation.Region)
	m.RoleArn = types.StringValue(athenaDto.AthenaInformation.RoleArn)
	m.S3OutputLocation = types.StringValue(athenaDto.AthenaInformation.S3OutputLocation)
	m.VpcUrl = types.StringPointerValue(athenaDto.AthenaInformation.VpcUrl)
	m.Workgroup = types.StringValue(athenaDto.AthenaInformation.Workgroup)
	m.Schedule = types.StringPointerValue(athenaDto.Schedule)
	return diag.Diagnostics{}
}
