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

type QuickSightParametersModel struct {
	AccountId   types.String `tfsdk:"account_id"`
	AwsRegion   types.String `tfsdk:"aws_region"`
	RoleArn     types.String `tfsdk:"role_arn"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m QuickSightParametersModel) SchemaSourceType() string {
	return "quicksight"
}

func (m QuickSightParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "Your AWS account ID",
				Required:    true,
			},
			"aws_region": schema.StringAttribute{
				Description: "Your AWS region",
				Required:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "The ARN for your QuickSight role",
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

func (m QuickSightParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_id":  types.StringType,
		"aws_region":  types.StringType,
		"role_arn":    types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
	}
}

func (m QuickSightParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	quickSightParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.QuickSight = quickSightParams
	return o, diag.Diagnostics{}
}

func (m QuickSightParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	quickSightInformation := sifflet.QuicksightInformation{
		AccountId: m.AccountId.ValueString(),
		AwsRegion: m.AwsRegion.ValueString(),
		RoleArn:   m.RoleArn.ValueString(),
	}

	quickSightCreateDto := &sifflet.PublicCreateQuicksightSourceV2Dto{
		Name:                  name,
		Timezone:              &timezone,
		Type:                  sifflet.PublicCreateQuicksightSourceV2DtoTypeQUICKSIGHT,
		QuicksightInformation: &quickSightInformation,
		Credentials:           m.Credentials.ValueStringPointer(),
		Schedule:              m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(quickSightCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create QuickSight source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m QuickSightParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	quickSightInformation := sifflet.QuicksightInformation{
		AccountId: m.AccountId.ValueString(),
		AwsRegion: m.AwsRegion.ValueString(),
		RoleArn:   m.RoleArn.ValueString(),
	}

	quickSightUpdateDto := &sifflet.PublicUpdateQuicksightSourceV2Dto{
		Name:                  &name,
		Timezone:              &timezone,
		Type:                  sifflet.PublicUpdateQuicksightSourceV2DtoTypeQUICKSIGHT,
		QuicksightInformation: quickSightInformation,
		Credentials:           m.Credentials.ValueString(),
		Schedule:              m.Schedule.ValueStringPointer(),
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(quickSightUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update QuickSight source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *QuickSightParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	quickSightDto := d.PublicGetQuicksightSourceV2Dto
	if quickSightDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read QuickSight source", "Source does not contain QuickSight params but was interpreted as a QuickSight source")}
	}

	m.AccountId = types.StringValue(quickSightDto.QuicksightInformation.AccountId)
	m.AwsRegion = types.StringValue(quickSightDto.QuicksightInformation.AwsRegion)
	m.RoleArn = types.StringValue(quickSightDto.QuicksightInformation.RoleArn)
	m.Credentials = types.StringPointerValue(quickSightDto.Credentials)
	m.Schedule = types.StringPointerValue(quickSightDto.Schedule)
	return diag.Diagnostics{}
}
