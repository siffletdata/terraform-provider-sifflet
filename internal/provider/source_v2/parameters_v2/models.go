// Package source contains the models and schemas used to represent sources in the provider.
// A dedicated package is required to handle the complexity of the "parameters" field in the source resource,
// whose schema depends on the source type (BigQuery, dbt, Airflow, ...).
// To add support for a new source type:
// - Create a new file in this package, named after the source type (e.g. "bigquery.go"), and implement the sourceParameters interface for the new source type.
// - Add the new source type to the allSourceTypes map, and update ParameterModel to include the new source type (including the [Empty] function).
// - Add Terraform acceptance tests in source_resource_test.go.
package parameters_v2

import (
	"context"
	"fmt"
	"strings"

	// "terraform-provider-sifflet/internal/clientv2/openapi"
	"terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// allSourceTypes is a map of factory functions that return a new instance of a sourceParameters implementation for each supported source type.
var allSourceTypes = map[string]func() SourceParameters{
	AirflowParametersModel{}.SchemaSourceType():    func() SourceParameters { return &AirflowParametersModel{} },
	AthenaParametersModel{}.SchemaSourceType():     func() SourceParameters { return &AthenaParametersModel{} },
	BigQueryParametersModel{}.SchemaSourceType():   func() SourceParameters { return &BigQueryParametersModel{} },
	DatabricksParametersModel{}.SchemaSourceType(): func() SourceParameters { return &DatabricksParametersModel{} },
	DbtParametersModel{}.SchemaSourceType():        func() SourceParameters { return &DbtParametersModel{} },
	DbtCloudParametersModel{}.SchemaSourceType():   func() SourceParameters { return &DbtCloudParametersModel{} },
	FivetranParametersModel{}.SchemaSourceType():   func() SourceParameters { return &FivetranParametersModel{} },
	LookerParametersModel{}.SchemaSourceType():     func() SourceParameters { return &LookerParametersModel{} },
	MssqlParametersModel{}.SchemaSourceType():      func() SourceParameters { return &MssqlParametersModel{} },
	MysqlParametersModel{}.SchemaSourceType():      func() SourceParameters { return &MysqlParametersModel{} },
	OracleParametersModel{}.SchemaSourceType():     func() SourceParameters { return &OracleParametersModel{} },
	PostgresqlParametersModel{}.SchemaSourceType(): func() SourceParameters { return &PostgresqlParametersModel{} },
	PowerBiParametersModel{}.SchemaSourceType():    func() SourceParameters { return &PowerBiParametersModel{} },
	QuickSightParametersModel{}.SchemaSourceType(): func() SourceParameters { return &QuickSightParametersModel{} },
	RedshiftParametersModel{}.SchemaSourceType():   func() SourceParameters { return &RedshiftParametersModel{} },
	SnowflakeParametersModel{}.SchemaSourceType():  func() SourceParameters { return &SnowflakeParametersModel{} },
	SynapseParametersModel{}.SchemaSourceType():    func() SourceParameters { return &SynapseParametersModel{} },
	TableauParametersModel{}.SchemaSourceType():    func() SourceParameters { return &TableauParametersModel{} },
}

// ParametersModel represents the parameters for a source, regardless of the source type.
// This model should only contains the parameters for a single source type, with all other fields set to a null object ([types.ObjectNull]).
// The SourceType field is not guaranteed to be set before [SetSourceType] is called.
//
// Design note: ParametersModel doesn't implement the interfaces CreatableModel, UpdatableModel and so on, because we need information from the sourceV2Model
// to create the DTOs needed for source creation and update (name and timezone which are not part of the parameters).
// Instead, we rely on the SourceParameters interface to handle the conversion between the model and the DTO.
// This results in code that's probably more complicated than needed - consider refactoring if working on this code.
type ParametersModel struct {
	SourceType types.String `tfsdk:"source_type"`
	Airflow    types.Object `tfsdk:"airflow"`
	Athena     types.Object `tfsdk:"athena"`
	BigQuery   types.Object `tfsdk:"bigquery"`
	Databricks types.Object `tfsdk:"databricks"`
	Dbt        types.Object `tfsdk:"dbt"`
	DbtCloud   types.Object `tfsdk:"dbt_cloud"`
	Fivetran   types.Object `tfsdk:"fivetran"`
	Looker     types.Object `tfsdk:"looker"`
	Mssql      types.Object `tfsdk:"mssql"`
	Mysql      types.Object `tfsdk:"mysql"`
	Oracle     types.Object `tfsdk:"oracle"`
	Postgresql types.Object `tfsdk:"postgresql"`
	PowerBi    types.Object `tfsdk:"power_bi"`
	QuickSight types.Object `tfsdk:"quicksight"`
	Redshift   types.Object `tfsdk:"redshift"`
	Snowflake  types.Object `tfsdk:"snowflake"`
	Synapse    types.Object `tfsdk:"synapse"`
	Tableau    types.Object `tfsdk:"tableau"`
}

func NewParametersModel() ParametersModel {
	return ParametersModel{
		SourceType: types.StringNull(),
		Airflow:    types.ObjectNull(AirflowParametersModel{}.AttributeTypes()),
		Athena:     types.ObjectNull(AthenaParametersModel{}.AttributeTypes()),
		BigQuery:   types.ObjectNull(BigQueryParametersModel{}.AttributeTypes()),
		Databricks: types.ObjectNull(DatabricksParametersModel{}.AttributeTypes()),
		Dbt:        types.ObjectNull(DbtParametersModel{}.AttributeTypes()),
		DbtCloud:   types.ObjectNull(DbtCloudParametersModel{}.AttributeTypes()),
		Fivetran:   types.ObjectNull(FivetranParametersModel{}.AttributeTypes()),
		Looker:     types.ObjectNull(LookerParametersModel{}.AttributeTypes()),
		Mssql:      types.ObjectNull(MssqlParametersModel{}.AttributeTypes()),
		Mysql:      types.ObjectNull(MysqlParametersModel{}.AttributeTypes()),
		Oracle:     types.ObjectNull(OracleParametersModel{}.AttributeTypes()),
		Postgresql: types.ObjectNull(PostgresqlParametersModel{}.AttributeTypes()),
		PowerBi:    types.ObjectNull(PowerBiParametersModel{}.AttributeTypes()),
		QuickSight: types.ObjectNull(QuickSightParametersModel{}.AttributeTypes()),
		Redshift:   types.ObjectNull(RedshiftParametersModel{}.AttributeTypes()),
		Snowflake:  types.ObjectNull(SnowflakeParametersModel{}.AttributeTypes()),
		Synapse:    types.ObjectNull(SynapseParametersModel{}.AttributeTypes()),
		Tableau:    types.ObjectNull(TableauParametersModel{}.AttributeTypes()),
	}
}

// SourceParameters represents the parameters for a source type.
// Each source type has different parameters (e.g BigQuery has project_id, dataset_id... while dbt has project_name, target...).
// This interface allows the rest of the code to manipulate source parameters without knowing the specifics of each source type.
//
// Design notes: this interface doesn't extend CreatableModel, UpdatableModel, etc. because we need information from the sourceV2Model
// to create the DTOs needed for source creation and update (name and timezone which are not part of the parameters). This results in
// slightly different code conventions in this package. Consider refactoring that if working on this code.
type SourceParameters interface {
	// TerraformSchema returns the Terraform resource schema for this source type.
	TerraformSchema() schema.SingleNestedAttribute

	// AttributeTypes returns the attribute types for this source type. Attribute types
	// are what Terraform use to convert the object stored in the state to a Go struct (and vice versa).
	// It must match the attributes defined in the Schema method.
	AttributeTypes() map[string]attr.Type

	// AsParametersModel creates a ParametersModel populated with the values from this source parameters.
	AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics)

	// ToCreateDto creates a DTO (data transfer object) that can be sent to the API to create a source.
	// This method takes in input the name and timezone of the source, which are not part of the parameters but are needed for the DTO.
	ToCreateDto(ctx context.Context, name string, timezone string) (client.PublicCreateSourceV2JSONBody, diag.Diagnostics)

	// ToUpdateDto creates a DTO (data transfer object) that can be sent to the API to update a source.
	// This method takes in input the name and timezone of the source, which are not part of the parameters but are needed for the DTO.
	ToUpdateDto(ctx context.Context, name string, timezone string) (client.PublicEditSourceV2JSONBody, diag.Diagnostics)

	// ModelFromDto populates the struct with the values from the given DTO.
	// This method will error if the given DTO type does not match this source type.
	ModelFromDto(ctx context.Context, d client.SiffletPublicGetSourceV2Dto) diag.Diagnostics

	// SchemaSourceType returns the source type as a string, as accepted by the Terraform schema (e.g "bigquery", in lowercase).
	SchemaSourceType() string
}

// ParamsImplFromSchemaName returns the SourceParameters implementation for the given source type.
// The source type name must match what's stored in the Terraform schema (e.g in lowercase).
func ParamsImplFromSchemaName(sourceType string) (SourceParameters, error) {
	if sourceParamsBuilder, ok := allSourceTypes[sourceType]; ok {
		return sourceParamsBuilder(), nil
	}
	return nil, fmt.Errorf("unknown source type %s", sourceType)
}

// ParamsImplFromApiResponseName returns the SourceParameters implementation for the given source type.
// The source type name must match what's returned by the API (e.g in uppercase).
func ParamsImplFromApiResponseName(apiSourceType string) (SourceParameters, error) {
	return ParamsImplFromSchemaName(strings.ToLower(apiSourceType))
}

// GetAllSourceTypes returns a list of all supported source types (as represented in the Terraform schema).
func GetAllSourceTypes() []string {
	var out []string
	for sourceType := range allSourceTypes {
		out = append(out, sourceType)
	}
	return out
}

func (m ParametersModel) AttributeTypes() map[string]attr.Type {
	var out = make(map[string]attr.Type)

	// The returned map looks like:
	// return map[string]attr.Type{
	// 	"bigquery": types.ObjectType{
	// 		AttrTypes: BigQueryParametersModel{}.AttributeTypes(),
	// 	},
	// 	"dbt": types.ObjectType{
	// 		AttrTypes: DbtParametersModel{}.AttributeTypes(),
	// 	},
	//  ...
	// 	"source_type": types.StringType,
	// }

	for _, factory := range allSourceTypes {
		t := factory()
		out[t.SchemaSourceType()] = types.ObjectType{
			AttrTypes: t.AttributeTypes(),
		}
	}
	out["source_type"] = types.StringType

	return out
}

func (m ParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	attributes := make(map[string]schema.Attribute)
	attributes["source_type"] = schema.StringAttribute{
		Description: "Source type (e.g bigquery, dbt, ...). This attribute is automatically set depending on which connection parameters are set.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			// The parent "parameters" attribute has a plan modifier that ensures that the source will be recreated
			// if the source type changes. Thus, if the source type doesn't change, we can keep the existing state value,
			// and if it does, the whole resource will be recreated anyway.
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	for _, factory := range allSourceTypes {
		t := factory()
		// A resource-level validator ensure that only one type of parameters is provided
		attributes[t.SchemaSourceType()] = t.TerraformSchema()
	}

	return schema.SingleNestedAttribute{
		Description: "Connection parameters. Provide only one nested block depending on the source type.",
		Required:    true,
		Attributes:  attributes,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplaceIf(
				func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
					// If the source type changes, the resource must be replaced (the API will reject a type change)
					resp.RequiresReplace = true

					var state ParametersModel
					var plan ParametersModel

					diags := req.State.GetAttribute(ctx, path.Root("parameters"), &state)
					if diags.HasError() {
						resp.Diagnostics.Append(diags...)
						return
					}

					diags = req.Plan.GetAttribute(ctx, path.Root("parameters"), &plan)
					if diags.HasError() {
						resp.Diagnostics.Append(diags...)
						return
					}

					previousSource, diags := state.GetSourceParameter(ctx)
					if diags.HasError() {
						resp.Diagnostics.Append(diags.Errors()...)
						return
					}
					nextSource, diags := plan.GetSourceParameter(ctx)
					if diags.HasError() {
						resp.Diagnostics.Append(diags.Errors()...)
						return
					}

					if previousSource.SchemaSourceType() == nextSource.SchemaSourceType() {
						resp.RequiresReplace = false
					}
				},
				"If the source type changes, the resource must be replaced.",
				"If the source type changes, the resource must be replaced.",
			),
		},
	}
}

func (m *ParametersModel) SetSourceType(ctx context.Context) diag.Diagnostics {
	t, diags := m.GetSourceParameter(ctx)
	if diags != nil {
		return diags
	}
	m.SourceType = types.StringValue(t.SchemaSourceType())
	return nil
}

// GetSourceParameter returns the SourceParameters implementation for the one source in the parameters block that is set in the ParametersModel.
func (m ParametersModel) GetSourceParameter(ctx context.Context) (SourceParameters, diag.Diagnostics) {
	// Define a slice of parameter fields to check
	parameterFields := []struct {
		value types.Object
		model func() SourceParameters
	}{
		{m.Airflow, func() SourceParameters { return &AirflowParametersModel{} }},
		{m.Athena, func() SourceParameters { return &AthenaParametersModel{} }},
		{m.BigQuery, func() SourceParameters { return &BigQueryParametersModel{} }},
		{m.Databricks, func() SourceParameters { return &DatabricksParametersModel{} }},
		{m.DbtCloud, func() SourceParameters { return &DbtCloudParametersModel{} }},
		{m.Dbt, func() SourceParameters { return &DbtParametersModel{} }},
		{m.Fivetran, func() SourceParameters { return &FivetranParametersModel{} }},
		{m.Looker, func() SourceParameters { return &LookerParametersModel{} }},
		{m.Mssql, func() SourceParameters { return &MssqlParametersModel{} }},
		{m.Mysql, func() SourceParameters { return &MysqlParametersModel{} }},
		{m.Oracle, func() SourceParameters { return &OracleParametersModel{} }},
		{m.Postgresql, func() SourceParameters { return &PostgresqlParametersModel{} }},
		{m.PowerBi, func() SourceParameters { return &PowerBiParametersModel{} }},
		{m.QuickSight, func() SourceParameters { return &QuickSightParametersModel{} }},
		{m.Redshift, func() SourceParameters { return &RedshiftParametersModel{} }},
		{m.Snowflake, func() SourceParameters { return &SnowflakeParametersModel{} }},
		{m.Synapse, func() SourceParameters { return &SynapseParametersModel{} }},
		{m.Tableau, func() SourceParameters { return &TableauParametersModel{} }},
	}

	// Iterate through the fields and return the first non-null one
	for _, field := range parameterFields {
		if !field.value.IsNull() {
			sourceParams := field.model()
			diags := field.value.As(ctx, sourceParams, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, diags
			}
			return sourceParams, nil
		}
	}

	return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Could not build source object", "Could not determine source type from the provided configuration (the parameters don't match any known type), this is a bug in the provider")}
}

func (m ParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (client.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	sourceParams, diags := m.GetSourceParameter(ctx)
	if diags.HasError() {
		return client.PublicCreateSourceV2JSONBody{}, diags
	}
	dto, diags := sourceParams.ToCreateDto(ctx, name, timezone)
	if diags.HasError() {
		return client.PublicCreateSourceV2JSONBody{}, diags
	}
	return dto, diag.Diagnostics{}
}

func (m ParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (client.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	sourceParams, diags := m.GetSourceParameter(ctx)
	if diags.HasError() {
		return client.PublicEditSourceV2JSONBody{}, diags
	}
	dto, diags := sourceParams.ToUpdateDto(ctx, name, timezone)
	if diags.HasError() {
		return client.PublicEditSourceV2JSONBody{}, diags
	}
	return dto, diag.Diagnostics{}
}

func SourceParametersModelFromDto(ctx context.Context, dto client.SiffletPublicGetSourceV2Dto) (SourceParameters, diag.Diagnostics) {
	source, err := dto.GetSourceDto()
	if err != nil {
		return nil, tfutils.ErrToDiags("Unable to read source", err)
	}
	sourceType := source.GetType()

	sourceTypeParams, err := ParamsImplFromApiResponseName(sourceType)
	if err != nil {
		return nil, tfutils.ErrToDiags("Unsupported source type", err)
	}

	diags := sourceTypeParams.ModelFromDto(ctx, dto)
	if diags.HasError() {
		return nil, diags
	}

	return sourceTypeParams, diag.Diagnostics{}
}
