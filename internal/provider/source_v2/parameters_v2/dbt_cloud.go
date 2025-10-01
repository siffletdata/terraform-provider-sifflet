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

type DbtCloudParametersModel struct {
	AccountId   types.String `tfsdk:"account_id"`
	BaseUrl     types.String `tfsdk:"base_url"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m DbtCloudParametersModel) SchemaSourceType() string {
	return "dbtcloud"
}

func (m DbtCloudParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "Your dbt Cloud account ID",
				Required:    true,
			},
			"base_url": schema.StringAttribute{
				Description: "Your dbt Cloud base URL",
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

func (m DbtCloudParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_id":  types.StringType,
		"base_url":    types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
	}
}

func (m DbtCloudParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	dbtCloudParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.DbtCloud = dbtCloudParams
	return o, diag.Diagnostics{}
}

func (m DbtCloudParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	dbtCloudInformation := sifflet.DbtCloudInformation{
		AccountId: m.AccountId.ValueString(),
		BaseUrl:   m.BaseUrl.ValueString(),
	}

	dbtCloudCreateDto := &sifflet.PublicCreateDbtCloudSourceV2Dto{
		Name:                name,
		Timezone:            &timezone,
		Type:                sifflet.PublicCreateDbtCloudSourceV2DtoTypeDBTCLOUD,
		DbtCloudInformation: &dbtCloudInformation,
		Credentials:         m.Credentials.ValueStringPointer(),
		Schedule:            m.Schedule.ValueStringPointer(),
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err := createSourceJsonBody.FromAny(dbtCloudCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create dbt Cloud source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m DbtCloudParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	dbtCloudInformation := sifflet.DbtCloudInformation{
		AccountId: m.AccountId.ValueString(),
		BaseUrl:   m.BaseUrl.ValueString(),
	}

	dbtCloudUpdateDto := &sifflet.PublicUpdateDbtCloudSourceV2Dto{
		Name:                &name,
		Timezone:            &timezone,
		Type:                sifflet.PublicUpdateDbtCloudSourceV2DtoTypeDBTCLOUD,
		DbtCloudInformation: dbtCloudInformation,
		Credentials:         m.Credentials.ValueString(),
		Schedule:            m.Schedule.ValueStringPointer(),
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err := editSourceJsonBody.FromAny(dbtCloudUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update dbt Cloud source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *DbtCloudParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	dbtCloudDto := d.PublicGetDbtCloudSourceV2Dto
	if dbtCloudDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read dbt Cloud source", "Source does not contain dbt Cloud params but was interpreted as a dbt Cloud source")}
	}

	m.AccountId = types.StringValue(dbtCloudDto.DbtCloudInformation.AccountId)
	m.BaseUrl = types.StringValue(dbtCloudDto.DbtCloudInformation.BaseUrl)
	m.Credentials = types.StringPointerValue(dbtCloudDto.Credentials)
	m.Schedule = types.StringPointerValue(dbtCloudDto.Schedule)
	return diag.Diagnostics{}
}
