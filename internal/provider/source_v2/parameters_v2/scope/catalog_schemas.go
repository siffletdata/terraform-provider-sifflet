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
	CatalogSchemasScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"catalogs": types.ListType{
				ElemType: catalogSchemasTypeAttributes,
			},
		},
	}
)

var (
	catalogSchemasTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"schemas": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type CatalogSchemas struct {
	Name    types.String `tfsdk:"name"`
	Schemas types.List   `tfsdk:"schemas"`
}

type CatalogSchemasScopeModel struct {
	Type     types.String `tfsdk:"type"`
	Catalogs types.List   `tfsdk:"catalogs"`
}

func parseCatalogSchemasScopeType(v string) (sifflet.PublicCatalogSchemasScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicCatalogSchemasScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicCatalogSchemasScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Catalog Schemas Scope type: %s", v)
	}
}

func catalogSchemasScopeTypeToString(v sifflet.PublicCatalogSchemasScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicCatalogSchemasScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicCatalogSchemasScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Catalog Schemas Scope type: %s", v)
	}
}

func ToPublicCatalogSchemasScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicCatalogSchemasScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeModel CatalogSchemasScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	catalogSchemas := make([]types.Object, 0, len(scopeModel.Catalogs.Elements()))
	diags = scopeModel.Catalogs.ElementsAs(ctx, &catalogSchemas, false)
	if diags.HasError() {
		return nil, diags
	}

	catalogSchemasDto := make([]sifflet.CatalogSchema, len(catalogSchemas))
	for i, catalogSchema := range catalogSchemas {
		var catalogSchemaModel CatalogSchemas
		diags = catalogSchema.As(ctx, &catalogSchemaModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		var schemas []string
		diags = catalogSchemaModel.Schemas.ElementsAs(ctx, &schemas, false)
		if diags.HasError() {
			return nil, diags
		}
		catalogSchemasDto[i] = sifflet.CatalogSchema{
			Name:    catalogSchemaModel.Name.ValueString(),
			Schemas: &schemas,
		}
	}

	scopeType, err := parseCatalogSchemasScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Catalog Schemas Scope Dto", err)
	}

	scopeDto := sifflet.PublicCatalogSchemasScopeDto{
		Type:     scopeType,
		Catalogs: catalogSchemasDto,
	}

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicCatalogSchemasScopeDto(ctx context.Context, scopeDto *sifflet.PublicCatalogSchemasScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(CatalogSchemasScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}

	scopeTypeString, err := catalogSchemasScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(CatalogSchemasScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Schemas Scope Dto", err)
	}

	catalogSchemasModel := make([]CatalogSchemas, len(scopeDto.Catalogs))
	for i, catalogSchemaDto := range scopeDto.Catalogs {
		schemas, diags := types.ListValueFrom(ctx, types.StringType, catalogSchemaDto.Schemas)
		if diags.HasError() {
			return types.ObjectNull(CatalogSchemasScopeTypeAttributes.AttrTypes), diags
		}
		catalogSchemasModel[i] = CatalogSchemas{
			Name:    types.StringValue(catalogSchemaDto.Name),
			Schemas: schemas,
		}
	}
	scopeCatalogSchemas, diags := types.ListValueFrom(ctx, catalogSchemasTypeAttributes, catalogSchemasModel)
	if diags.HasError() {
		return types.ObjectNull(CatalogSchemasScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel := CatalogSchemasScopeModel{
		Type:     types.StringValue(scopeTypeString),
		Catalogs: scopeCatalogSchemas,
	}
	scopeObject, diags := types.ObjectValueFrom(ctx, CatalogSchemasScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(CatalogSchemasScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
