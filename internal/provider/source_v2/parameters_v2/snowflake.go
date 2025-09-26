package parameters_v2

import (
	"context"
	"encoding/json"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/source_v2/parameters_v2/scope"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SnowflakeParametersModel struct {
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	Warehouse         types.String `tfsdk:"warehouse"`
	Credentials       types.String `tfsdk:"credentials"`
	Scope             types.Object `tfsdk:"scope"`
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
			"scope": schema.SingleNestedAttribute{
				Description: "Database schemas to include or exclude. If not specified, all the database schemas will be included (including future ones).",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Whether to include or exclude the specified database schemas. One of INCLUSION or EXCLUSION.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INCLUSION", "EXCLUSION"),
						},
					},
					"databases": schema.ListNestedAttribute{
						Description: "The database schemas to either include or exclude.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Database name",
									Required:    true,
								},
								"schemas": schema.ListAttribute{
									ElementType: types.StringType,
									Required:    true,
									Description: "List of schema names within this database",
								},
							},
						},
					},
				},
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
		"scope":              scope.DatabaseSchemasScopeTypeAttributes,
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

func (m SnowflakeParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	snowflakeInformation := sifflet.SnowflakeInformation{
		AccountIdentifier: m.AccountIdentifier.ValueString(),
		Warehouse:         m.Warehouse.ValueString(),
	}

	scopeDto, diags := scope.ToPublicDatabasesSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	snowflakeCreateDto := &sifflet.PublicCreateSnowflakeSourceV2Dto{
		Name:                 name,
		Timezone:             &timezone,
		Type:                 sifflet.PublicCreateSnowflakeSourceV2DtoTypeSNOWFLAKE,
		SnowflakeInformation: &snowflakeInformation,
		Credentials:          m.Credentials.ValueStringPointer(),
		Schedule:             m.Schedule.ValueStringPointer(),
		Scope:                scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(snowflakeCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Snowflake source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m SnowflakeParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	snowflakeInformation := sifflet.SnowflakeInformation{
		AccountIdentifier: m.AccountIdentifier.ValueString(),
		Warehouse:         m.Warehouse.ValueString(),
	}

	scopeDto, diags := scope.ToPublicDatabasesSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	snowflakeUpdateDto := &sifflet.PublicUpdateSnowflakeSourceV2Dto{
		Name:                 &name,
		Timezone:             &timezone,
		Type:                 sifflet.PublicUpdateSnowflakeSourceV2DtoTypeSNOWFLAKE,
		SnowflakeInformation: snowflakeInformation,
		Credentials:          m.Credentials.ValueString(),
		Schedule:             m.Schedule.ValueStringPointer(),
		Scope:                scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(snowflakeUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Snowflake source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

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
	scopeObject, diags := scope.FromPublicDatabasesSchemasScopeDto(ctx, snowflakeDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}
