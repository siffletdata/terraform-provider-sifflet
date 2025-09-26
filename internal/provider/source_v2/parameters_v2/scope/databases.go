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
	DatabasesScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"databases": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type DatabasesScopeModel struct {
	Type      types.String `tfsdk:"type"`
	Databases types.List   `tfsdk:"databases"`
}

func parseDatabasesScopeType(v string) (sifflet.PublicDatabasesScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicDatabasesScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicDatabasesScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Databases Scope type: %s", v)
	}
}

func databasesScopeTypeToString(v sifflet.PublicDatabasesScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicDatabasesScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicDatabasesScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Databases Scope type: %s", v)
	}
}

func ToPublicDatabasesScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicDatabasesScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeDto sifflet.PublicDatabasesScopeDto
	var scopeModel DatabasesScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	var databases []string
	diags = scopeModel.Databases.ElementsAs(ctx, &databases, false)
	if diags.HasError() {
		return nil, diags
	}

	scopeType, err := parseDatabasesScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Databases Scope Dto", err)
	}
	scopeDto.Databases = &databases
	scopeDto.Type = scopeType

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicDatabasesScopeDto(ctx context.Context, scopeDto *sifflet.PublicDatabasesScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(DatabasesScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}
	var scopeModel DatabasesScopeModel
	scopeTypeString, err := databasesScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(DatabasesScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Databases Scope Dto", err)
	}
	scopeDatabases, diags := types.ListValueFrom(ctx, types.StringType, scopeDto.Databases)
	if diags.HasError() {
		return types.ObjectNull(DatabasesScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel.Type = types.StringValue(scopeTypeString)
	scopeModel.Databases = scopeDatabases
	scopeObject, diags := types.ObjectValueFrom(ctx, DatabasesScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(DatabasesScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
