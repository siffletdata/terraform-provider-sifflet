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
	FoldersScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"folders": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type FoldersScopeModel struct {
	Type    types.String `tfsdk:"type"`
	Folders types.List   `tfsdk:"folders"`
}

func parseFoldersScopeType(v string) (sifflet.PublicFoldersScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicFoldersScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicFoldersScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Folders Scope type: %s", v)
	}
}

func foldersScopeTypeToString(v sifflet.PublicFoldersScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicFoldersScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicFoldersScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Folders Scope type: %s", v)
	}
}

func ToPublicFoldersScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicFoldersScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeDto sifflet.PublicFoldersScopeDto
	var scopeModel FoldersScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	var folders []string
	diags = scopeModel.Folders.ElementsAs(ctx, &folders, false)
	if diags.HasError() {
		return nil, diags
	}

	scopeType, err := parseFoldersScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Folders Scope Dto", err)
	}
	scopeDto.Folders = &folders
	scopeDto.Type = scopeType

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicFoldersScopeDto(ctx context.Context, scopeDto *sifflet.PublicFoldersScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(FoldersScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}
	var scopeModel FoldersScopeModel
	scopeTypeString, err := foldersScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(FoldersScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Folders Scope Dto", err)
	}
	scopeFolders, diags := types.ListValueFrom(ctx, types.StringType, scopeDto.Folders)
	if diags.HasError() {
		return types.ObjectNull(FoldersScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel.Type = types.StringValue(scopeTypeString)
	scopeModel.Folders = scopeFolders
	scopeObject, diags := types.ObjectValueFrom(ctx, FoldersScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(FoldersScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
