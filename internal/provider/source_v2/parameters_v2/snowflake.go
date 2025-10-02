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

type SnowflakeParametersModel struct {
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	Warehouse         types.String `tfsdk:"warehouse"`
	Credentials       types.String `tfsdk:"credentials"`
	Schedule          types.String `tfsdk:"schedule"`
}

func (m SnowflakeParametersModel) SchemaSourceType() string {
	return "snowflake"
}

func (m SnowflakeParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"account_identifier": schema.StringAttribute{
				Description: "Snowflake account identifier",
				Required:    true,
			},
			"warehouse": schema.StringAttribute{
				Description: "Snowflake warehouse name",
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

func (m SnowflakeParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_identifier": types.StringType,
		"warehouse":          types.StringType,
		"credentials":        types.StringType,
		"schedule":           types.StringType,
	}
}

func (m SnowflakeParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	snowflakeParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Snowflake = snowflakeParams
	return o, diag.Diagnostics{}
}

func (m SnowflakeParametersModel) ToCreateDto(ctx context.Context, name string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	snowflakeInformation := sifflet.SnowflakeInformation{
		AccountIdentifier: m.AccountIdentifier.ValueString(),
		Warehouse:         m.Warehouse.ValueString(),
	}

	snowflakeCreateDto := &sifflet.PublicCreateSnowflakeSourceV2Dto{
		Name:                 name,
		Type:                 sifflet.PublicCreateSnowflakeSourceV2DtoTypeSNOWFLAKE,
		SnowflakeInformation: &snowflakeInformation,
		Credentials:          m.Credentials.ValueStringPointer(),
		Schedule:             m.Schedule.ValueStringPointer(),
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err := createSourceJsonBody.FromAny(snowflakeCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Snowflake source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m SnowflakeParametersModel) ToUpdateDto(ctx context.Context, name string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	snowflakeInformation := sifflet.SnowflakeInformation{
		AccountIdentifier: m.AccountIdentifier.ValueString(),
		Warehouse:         m.Warehouse.ValueString(),
	}

	snowflakeUpdateDto := &sifflet.PublicUpdateSnowflakeSourceV2Dto{
		Name:                 &name,
		Type:                 sifflet.PublicUpdateSnowflakeSourceV2DtoTypeSNOWFLAKE,
		SnowflakeInformation: snowflakeInformation,
		Credentials:          m.Credentials.ValueString(),
		Schedule:             m.Schedule.ValueStringPointer(),
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err := editSourceJsonBody.FromAny(snowflakeUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Snowflake source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *SnowflakeParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	snowflakeDto := d.PublicGetSnowflakeSourceV2Dto
	if snowflakeDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Snowflake source", "Source does not contain Snowflake params but was interpreted as a Snowflake source")}
	}

	m.AccountIdentifier = types.StringValue(snowflakeDto.SnowflakeInformation.AccountIdentifier)
	m.Warehouse = types.StringValue(snowflakeDto.SnowflakeInformation.Warehouse)
	m.Credentials = types.StringPointerValue(snowflakeDto.Credentials)
	m.Schedule = types.StringPointerValue(snowflakeDto.Schedule)
	return diag.Diagnostics{}
}
