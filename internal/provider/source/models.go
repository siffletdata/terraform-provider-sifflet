// Package source contains the models and schemas used to represent sources in the provider.
// A dedicated package is required to handle the complexity of the "parameters" field in the source resource,
// whose schema depends on the source type (BigQuery, dbt, Airflow, ...).
// To add support for a new source type:
// - Create a new file in this package, named after the source type (e.g. "bigquery.go"), and implement the sourceParameters interface for the new source type.
// - Add the new source type to the allSourceTypes map, and update ParameterModel to include the new source type (including the [Empty] function).
// - Add Terraform acceptance tests in source_resource_test.go.
package source

import (
	"context"
	"fmt"
	"strings"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var allSourceTypes = map[string]sourceParameters{
	AirflowParametersModel{}.SchemaSourceType():    &AirflowParametersModel{},
	AthenaParametersModel{}.SchemaSourceType():     &AthenaParametersModel{},
	BigQueryParametersModel{}.SchemaSourceType():   &BigQueryParametersModel{},
	DatabricksParametersModel{}.SchemaSourceType(): &DatabricksParametersModel{},
	DbtParametersModel{}.SchemaSourceType():        &DbtParametersModel{},
	DbtCloudParametersModel{}.SchemaSourceType():   &DbtCloudParametersModel{},
	FivetranParametersModel{}.SchemaSourceType():   &FivetranParametersModel{},
	HiveParametersModel{}.SchemaSourceType():       &HiveParametersModel{},
	LookerParametersModel{}.SchemaSourceType():     &LookerParametersModel{},
	MssqlParametersModel{}.SchemaSourceType():      &MssqlParametersModel{},
	MysqlParametersModel{}.SchemaSourceType():      &MysqlParametersModel{},
	OracleParametersModel{}.SchemaSourceType():     &OracleParametersModel{},
	PostgresqlParametersModel{}.SchemaSourceType(): &PostgresqlParametersModel{},
	PowerBiParametersModel{}.SchemaSourceType():    &PowerBiParametersModel{},
	QuickSightParametersModel{}.SchemaSourceType(): &QuickSightParametersModel{},
	RedshiftParametersModel{}.SchemaSourceType():   &RedshiftParametersModel{},
	SnowflakeParametersModel{}.SchemaSourceType():  &SnowflakeParametersModel{},
	SynapseParametersModel{}.SchemaSourceType():    &SynapseParametersModel{},
	TableauParametersModel{}.SchemaSourceType():    &TableauParametersModel{},
}

// ParametersModel represents the parameters for a source, regardless of the source type.
// This model should only contains the parameters for a single source type, with all other fields set to a null object ([types.ObjectNull]).
// The SourceType field is not guaranteed to be set before [SetSourceType] is called.
type ParametersModel struct {
	SourceType types.String `tfsdk:"source_type"`
	Airflow    types.Object `tfsdk:"airflow"`
	Athena     types.Object `tfsdk:"athena"`
	BigQuery   types.Object `tfsdk:"bigquery"`
	Databricks types.Object `tfsdk:"databricks"`
	Dbt        types.Object `tfsdk:"dbt"`
	DbtCloud   types.Object `tfsdk:"dbt_cloud"`
	Fivetran   types.Object `tfsdk:"fivetran"`
	Hive       types.Object `tfsdk:"hive"`
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
		Hive:       types.ObjectNull(HiveParametersModel{}.AttributeTypes()),
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

// sourceParameters represents the parameters for a source type.
// Each source type has different parameters (e.g BigQuery has project_id, dataset_id... while dbt has project_name, target...).
// This interface allows the rest of the code to manipulate source parameters without knowing the specifics of each source type.
type sourceParameters interface {
	// TerraformSchema returns the Terraform resource schema for this source type.
	TerraformSchema() schema.SingleNestedAttribute

	// AttributeTypes returns the attribute types for this source type. Attribute types
	// are what Terraform use to convert the object stored in the state to a Go struct (and vice versa).
	// It must match the attributes defined in the Schema method.
	AttributeTypes() map[string]attr.Type

	// AsParametersModel creates a ParametersModel populated with the values from this source parameters.
	AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics)

	// IsRepresentedBy returns true if the given ParametersModel contains Parameters
	// that matches this source type. The implementation must not use the SourceType
	// field of the ParametersModel to determine this, as it may not be set when
	// this method is called.
	IsRepresentedBy(ParametersModel) bool

	// DtoFromModel populates the struct with the values from the given ParametersModel, then
	// converts the parameters models to a DTO (data transfer object) that can be sent to the API.
	// This method may assume that the given ParametersModel type matches this source type.
	DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics)

	// ModelFromDto populates the struct with the values from the given DTO.
	// This method may assume that the given DTO type matches this source type.
	ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics

	// RequiresCredential returns true if this source type needs credentials.
	RequiresCredential() bool

	// SchemaSourceType returns the source type as a string, as accepted by the Terraform schema (e.g "bigquery", in lowercase).
	SchemaSourceType() string
}

// ApiSourceType returns the source type as a string, as accepted by the API (e.g "BIGQUERY", in uppercase).
func ApiSourceType(p sourceParameters) string {
	return strings.ToUpper(p.SchemaSourceType())
}

// ParamsImplFromSchemaName returns the sourceParameters implementation for the given source type.
// The source type name must match what's stored in the Terraform schema (e.g in lowercase).
func ParamsImplFromSchemaName(sourceType string) (sourceParameters, error) {
	if sourceParams, ok := allSourceTypes[sourceType]; ok {
		return sourceParams, nil
	}
	return nil, fmt.Errorf("Unknown source type %s", sourceType)
}

// ParamsImplFromApiResponseName returns the sourceParameters implementation for the given source type.
// The source type name must match what's returned by the API (e.g in uppercase).
func ParamsImplFromApiResponseName(apiSourceType string) (sourceParameters, error) {
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

	for _, t := range allSourceTypes {
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
		Description: "Source type (e.g BIGQUERY, DBT, ...). This attribute is automatically set depending on which connection parameters are set.",
		Computed:    true,
	}
	for _, t := range allSourceTypes {
		// A resource-level validator ensure that only one type of parameters is provided
		attributes[t.SchemaSourceType()] = t.TerraformSchema()
	}

	return schema.SingleNestedAttribute{
		Description: "Connection parameters. Provide only one nested block depending on the source type.",
		Required:    true,
		PlanModifiers: []planmodifier.Object{
			// The API doesn't allow yet to update parameters (or this behaviour is not correctly documented), see PLTE-964. In the meantime, we'll replace the datasource if the parameters change.
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: attributes,
	}
}

func (m *ParametersModel) SetSourceType() error {
	t, err := m.GetSourceType()
	if err != nil {
		return err
	}
	m.SourceType = types.StringValue(t.SchemaSourceType())
	return nil
}

func (m ParametersModel) GetSourceType() (sourceParameters, error) {
	for _, sourceParams := range allSourceTypes {
		if sourceParams.IsRepresentedBy(m) {
			return sourceParams, nil
		}
	}
	return nil, fmt.Errorf("Could not determine source type from the provided configuration (the parameters don't match any known type). This is a bug in the provider.")
}

type SourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Credential  types.String `tfsdk:"credential"`
	Schedule    types.String `tfsdk:"schedule"`
	Timezone    types.String `tfsdk:"timezone"`
	Parameters  types.Object `tfsdk:"parameters"`
	Tags        types.List   `tfsdk:"tags"`
}

func (m SourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"credential":  types.StringType,
		"schedule":    types.StringType,
		"timezone":    types.StringType,
		"parameters": types.ObjectType{
			AttrTypes: ParametersModel{}.AttributeTypes(),
		},
		"tags": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: TagModel{}.AttributeTypes(),
			},
		},
	}
}

type TagModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
	Kind types.String `tfsdk:"kind"`
}

func (m TagModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
		"kind": types.StringType,
	}
}

func ParseTagKind(s string) (sifflet.PublicTagReferenceDtoKind, error) {
	switch s {
	case "Tag":
		return sifflet.Tag, nil
	case "Classification":
		return sifflet.Classification, nil
	default:
		return "", fmt.Errorf("unsupported tag kind: %s", s)
	}
}

func TagKindToString(t sifflet.PublicTagReferenceDtoKind) (string, error) {
	switch t {
	case sifflet.Tag:
		return "Tag", nil
	case sifflet.Classification:
		return "Classification", nil
	default:
		return "", fmt.Errorf("unsupported tag kind: %s", t)
	}
}
