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
	SchemasScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"schemas": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type SchemasScopeModel struct {
	Type    types.String `tfsdk:"type"`
	Schemas types.List   `tfsdk:"schemas"`
}

func parseSchemasScopeType(v string) (sifflet.PublicSchemasScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicSchemasScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicSchemasScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Schemas Scope type: %s", v)
	}
}

func schemasScopeTypeToString(v sifflet.PublicSchemasScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicSchemasScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicSchemasScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Schemas Scope type: %s", v)
	}
}

func ToPublicSchemasScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicSchemasScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeDto sifflet.PublicSchemasScopeDto
	var scopeModel SchemasScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	var schemas []string
	diags = scopeModel.Schemas.ElementsAs(ctx, &schemas, false)
	if diags.HasError() {
		return nil, diags
	}

	scopeType, err := parseSchemasScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Schemas Scope Dto", err)
	}
	scopeDto.Schemas = &schemas
	scopeDto.Type = scopeType

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicSchemasScopeDto(ctx context.Context, scopeDto *sifflet.PublicSchemasScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(SchemasScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}
	var scopeModel SchemasScopeModel
	scopeTypeString, err := schemasScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(SchemasScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Schemas Scope Dto", err)
	}
	scopeSchemas, diags := types.ListValueFrom(ctx, types.StringType, scopeDto.Schemas)
	if diags.HasError() {
		return types.ObjectNull(SchemasScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel.Type = types.StringValue(scopeTypeString)
	scopeModel.Schemas = scopeSchemas
	scopeObject, diags := types.ObjectValueFrom(ctx, SchemasScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(SchemasScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
