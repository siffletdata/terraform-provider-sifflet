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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabricksParametersModel struct {
	Host        types.String `tfsdk:"host"`
	HttpPath    types.String `tfsdk:"http_path"`
	Port        types.Int32  `tfsdk:"port"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
	Scope       types.Object `tfsdk:"scope"`
}

func (m DatabricksParametersModel) SchemaSourceType() string {
	return "databricks"
}

func (m DatabricksParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Databricks host",
				Required:    true,
			},
			"http_path": schema.StringAttribute{
				Description: "Databricks HTTP path",
				Required:    true,
			},
			"port": schema.Int32Attribute{
				Description: "Databricks server port",
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
				Description: "Catalogs and schemas to include or exclude. If not specified, all the catalogs and schemas will be included (including future ones).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Whether to include or exclude the specified catalogs and schemas. One of INCLUSION or EXCLUSION.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INCLUSION", "EXCLUSION"),
						},
					},
					"catalogs": schema.ListNestedAttribute{
						Required:    true,
						Description: "The catalogs to either include or exclude.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of the catalog.",
									Required:    true,
								},
								"schemas": schema.ListAttribute{
									ElementType: types.StringType,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (m DatabricksParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"http_path":   types.StringType,
		"port":        types.Int32Type,
		"credentials": types.StringType,
		"schedule":    types.StringType,
		"scope":       scope.CatalogSchemasScopeTypeAttributes,
	}
}

func (m DatabricksParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	databricksParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Databricks = databricksParams
	return o, diag.Diagnostics{}
}

func (m DatabricksParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	databricksInformation := sifflet.DatabricksInformation{
		Host:     m.Host.ValueString(),
		HttpPath: m.HttpPath.ValueString(),
		Port:     m.Port.ValueInt32(),
	}

	scopeDto, diags := scope.ToPublicCatalogSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	databricksCreateDto := &sifflet.PublicCreateDatabricksSourceV2Dto{
		Name:                  name,
		Timezone:              &timezone,
		Type:                  sifflet.PublicCreateDatabricksSourceV2DtoTypeDATABRICKS,
		DatabricksInformation: &databricksInformation,
		Credentials:           m.Credentials.ValueStringPointer(),
		Schedule:              m.Schedule.ValueStringPointer(),
		Scope:                 scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(databricksCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Databricks source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m DatabricksParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	databricksInformation := sifflet.DatabricksInformation{
		Host:     m.Host.ValueString(),
		HttpPath: m.HttpPath.ValueString(),
		Port:     m.Port.ValueInt32(),
	}

	scopeDto, diags := scope.ToPublicCatalogSchemasScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	databricksUpdateDto := &sifflet.PublicUpdateDatabricksSourceV2Dto{
		Name:                  &name,
		Timezone:              &timezone,
		Type:                  sifflet.PublicUpdateDatabricksSourceV2DtoTypeDATABRICKS,
		DatabricksInformation: databricksInformation,
		Credentials:           m.Credentials.ValueString(),
		Schedule:              m.Schedule.ValueStringPointer(),
		Scope:                 scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(databricksUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Databricks source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *DatabricksParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	databricksDto := d.PublicGetDatabricksSourceV2Dto
	if databricksDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Databricks source", "Source does not contain Databricks params but was interpreted as a Databricks source")}
	}

	m.Host = types.StringValue(databricksDto.DatabricksInformation.Host)
	m.HttpPath = types.StringValue(databricksDto.DatabricksInformation.HttpPath)
	m.Port = types.Int32Value(databricksDto.DatabricksInformation.Port)
	m.Credentials = types.StringPointerValue(databricksDto.Credentials)
	m.Schedule = types.StringPointerValue(databricksDto.Schedule)
	scopeObject, diags := scope.FromPublicCatalogSchemasScopeDto(ctx, databricksDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}
