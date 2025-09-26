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
	ProjectsScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"projects": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type ProjectsScopeModel struct {
	Type     types.String `tfsdk:"type"`
	Projects types.List   `tfsdk:"projects"`
}

func parseProjectsScopeType(v string) (sifflet.PublicProjectsScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicProjectsScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicProjectsScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Projects Scope type: %s", v)
	}
}

func projectsScopeTypeToString(v sifflet.PublicProjectsScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicProjectsScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicProjectsScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Projects Scope type: %s", v)
	}
}

func ToPublicProjectsScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicProjectsScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeDto sifflet.PublicProjectsScopeDto
	var scopeModel ProjectsScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	var projects []string
	diags = scopeModel.Projects.ElementsAs(ctx, &projects, false)
	if diags.HasError() {
		return nil, diags
	}

	scopeType, err := parseProjectsScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Projects Scope Dto", err)
	}
	scopeDto.Projects = &projects
	scopeDto.Type = scopeType

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicProjectsScopeDto(ctx context.Context, scopeDto *sifflet.PublicProjectsScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(ProjectsScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}
	var scopeModel ProjectsScopeModel
	scopeTypeString, err := projectsScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(ProjectsScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Projects Scope Dto", err)
	}
	scopeProjects, diags := types.ListValueFrom(ctx, types.StringType, scopeDto.Projects)
	if diags.HasError() {
		return types.ObjectNull(ProjectsScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel.Type = types.StringValue(scopeTypeString)
	scopeModel.Projects = scopeProjects
	scopeObject, diags := types.ObjectValueFrom(ctx, ProjectsScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(ProjectsScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
