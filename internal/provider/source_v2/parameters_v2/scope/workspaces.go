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
	WorkspacesScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"workspaces": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type WorkspacesScopeModel struct {
	Type       types.String `tfsdk:"type"`
	Workspaces types.List   `tfsdk:"workspaces"`
}

func parseWorkspacesScopeType(v string) (sifflet.PublicWorkspacesScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.INCLUSION, nil
	case "EXCLUSION":
		return sifflet.EXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Workspaces Scope type: %s", v)
	}
}

func workspacesScopeTypeToString(v sifflet.PublicWorkspacesScopeDtoType) (string, error) {
	switch v {
	case sifflet.INCLUSION:
		return "INCLUSION", nil
	case sifflet.EXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Workspaces Scope type: %s", v)
	}
}

func ToPublicWorkspacesScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicWorkspacesScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeDto sifflet.PublicWorkspacesScopeDto
	var scopeModel WorkspacesScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	var workspaces []string
	diags = scopeModel.Workspaces.ElementsAs(ctx, &workspaces, false)
	if diags.HasError() {
		return nil, diags
	}

	scopeType, err := parseWorkspacesScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Workspaces Scope Dto", err)
	}
	scopeDto.Workspaces = &workspaces
	scopeDto.Type = scopeType

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicWorkspacesScopeDto(ctx context.Context, scopeDto *sifflet.PublicWorkspacesScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(WorkspacesScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}
	var scopeModel WorkspacesScopeModel
	scopeTypeString, err := workspacesScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(WorkspacesScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Workspaces Scope Dto", err)
	}
	scopeWorkspaces, diags := types.ListValueFrom(ctx, types.StringType, scopeDto.Workspaces)
	if diags.HasError() {
		return types.ObjectNull(WorkspacesScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel.Type = types.StringValue(scopeTypeString)
	scopeModel.Workspaces = scopeWorkspaces
	scopeObject, diags := types.ObjectValueFrom(ctx, WorkspacesScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(WorkspacesScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
