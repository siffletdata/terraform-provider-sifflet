package scope

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"
)

var (
	DatabaseSchemasScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"databases": types.ListType{
				ElemType: databaseSchemasTypeAttributes,
			},
		},
	}
)

var (
	databaseSchemasTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"schemas": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type DatabaseSchemas struct {
	Name    types.String `tfsdk:"name"`
	Schemas types.List   `tfsdk:"schemas"`
}

type DatabaseSchemasScopeModel struct {
	Type      types.String `tfsdk:"type"`
	Databases types.List   `tfsdk:"databases"`
}

func parseDatabasesSchemasScopeType(v string) (sifflet.PublicDatabaseSchemasScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicDatabaseSchemasScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicDatabaseSchemasScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Database Schemas Scope type: %s", v)
	}
}

func databasesSchemasScopeTypeToString(v sifflet.PublicDatabaseSchemasScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicDatabaseSchemasScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicDatabaseSchemasScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Database Schemas Scope type: %s", v)
	}
}

func ToPublicDatabasesSchemasScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicDatabaseSchemasScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeModel DatabaseSchemasScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	databaseSchemas := make([]types.Object, 0, len(scopeModel.Databases.Elements()))
	diags = scopeModel.Databases.ElementsAs(ctx, &databaseSchemas, false)
	if diags.HasError() {
		return nil, diags
	}

	databaseSchemasDto := make([]sifflet.DatabaseSchema, len(databaseSchemas))
	for i, databaseSchema := range databaseSchemas {
		var databaseSchemaModel DatabaseSchemas
		diags = databaseSchema.As(ctx, &databaseSchemaModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		var schemas []string
		diags = databaseSchemaModel.Schemas.ElementsAs(ctx, &schemas, false)
		if diags.HasError() {
			return nil, diags
		}
		databaseSchemasDto[i] = sifflet.DatabaseSchema{
			Name:    databaseSchemaModel.Name.ValueString(),
			Schemas: &schemas,
		}
	}

	scopeType, err := parseDatabasesSchemasScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Database Schemas Scope Dto", err)
	}

	scopeDto := sifflet.PublicDatabaseSchemasScopeDto{
		Type:      scopeType,
		Databases: databaseSchemasDto,
	}

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicDatabasesSchemasScopeDto(ctx context.Context, scopeDto *sifflet.PublicDatabaseSchemasScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(DatabaseSchemasScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}

	scopeTypeString, err := databasesSchemasScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(DatabaseSchemasScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Database Schemas Scope Dto", err)
	}

	databaseSchemasModel := make([]DatabaseSchemas, len(scopeDto.Databases))
	for i, databaseSchemaDto := range scopeDto.Databases {
		schemas, diags := types.ListValueFrom(ctx, types.StringType, databaseSchemaDto.Schemas)
		if diags.HasError() {
			return types.ObjectNull(DatabaseSchemasScopeTypeAttributes.AttrTypes), diags
		}
		databaseSchemasModel[i] = DatabaseSchemas{
			Name:    types.StringValue(databaseSchemaDto.Name),
			Schemas: schemas,
		}
	}
	scopeDatabaseSchemas, diags := types.ListValueFrom(ctx, databaseSchemasTypeAttributes, databaseSchemasModel)
	if diags.HasError() {
		return types.ObjectNull(DatabaseSchemasScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel := DatabaseSchemasScopeModel{
		Type:      types.StringValue(scopeTypeString),
		Databases: scopeDatabaseSchemas,
	}
	scopeObject, diags := types.ObjectValueFrom(ctx, DatabaseSchemasScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(DatabaseSchemasScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
