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

type SynapseParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Port        types.Int32  `tfsdk:"port"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
	Scope       types.Object `tfsdk:"scope"`
}

func (m SynapseParametersModel) SchemaSourceType() string {
	return "synapse"
}

func (m SynapseParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The host of your Synapse server",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Your Synapse server port",
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

func (m SynapseParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"port":        types.Int32Type,
		"credentials": types.StringType,
		"schedule":    types.StringType,
		"scope":       scope.DatabaseSchemasScopeTypeAttributes,
	}
}

func (m SynapseParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	synapseParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Synapse = synapseParams
	return o, diag.Diagnostics{}
}

func (m SynapseParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	synapseInformation := sifflet.SynapseInformation{
		Host: m.Host.ValueString(),
		Port: m.Port.ValueInt32(),
	}

	scopeDto, diags := scope.ToPublicDatabasesSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	synapseCreateDto := &sifflet.PublicCreateSynapseSourceV2Dto{
		Name:               name,
		Timezone:           &timezone,
		Type:               sifflet.PublicCreateSynapseSourceV2DtoTypeSYNAPSE,
		SynapseInformation: &synapseInformation,
		Credentials:        m.Credentials.ValueStringPointer(),
		Schedule:           m.Schedule.ValueStringPointer(),
		Scope:              scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(synapseCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Synapse source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m SynapseParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	synapseInformation := sifflet.SynapseInformation{
		Host: m.Host.ValueString(),
		Port: m.Port.ValueInt32(),
	}

	scopeDto, diags := scope.ToPublicDatabasesSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	synapseUpdateDto := &sifflet.PublicUpdateSynapseSourceV2Dto{
		Name:               &name,
		Timezone:           &timezone,
		Type:               sifflet.PublicUpdateSynapseSourceV2DtoTypeSYNAPSE,
		SynapseInformation: synapseInformation,
		Credentials:        m.Credentials.ValueString(),
		Schedule:           m.Schedule.ValueStringPointer(),
		Scope:              scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(synapseUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Synapse source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *SynapseParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	synapseDto := d.PublicGetSynapseSourceV2Dto
	if synapseDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Synapse source", "Source does not contain Synapse params but was interpreted as a Synapse source")}
	}

	m.Host = types.StringValue(synapseDto.SynapseInformation.Host)
	m.Port = types.Int32Value(synapseDto.SynapseInformation.Port)
	m.Credentials = types.StringPointerValue(synapseDto.Credentials)
	m.Schedule = types.StringPointerValue(synapseDto.Schedule)
	scopeObject, diags := scope.FromPublicDatabasesSchemasScopeDto(ctx, synapseDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}
