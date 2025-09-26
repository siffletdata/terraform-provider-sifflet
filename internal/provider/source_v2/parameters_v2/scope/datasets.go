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
	DatasetsScopeTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"datasets": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
)

type DatasetsScopeModel struct {
	Type     types.String `tfsdk:"type"`
	Datasets types.List   `tfsdk:"datasets"`
}

func parseDatasetsScopeType(v string) (sifflet.PublicDatasetsScopeDtoType, error) {
	switch v {
	case "INCLUSION":
		return sifflet.PublicDatasetsScopeDtoTypeINCLUSION, nil
	case "EXCLUSION":
		return sifflet.PublicDatasetsScopeDtoTypeEXCLUSION, nil
	default:
		return "", fmt.Errorf("unsupported value for Datasets Scope type: %s", v)
	}
}

func datasetsScopeTypeToString(v sifflet.PublicDatasetsScopeDtoType) (string, error) {
	switch v {
	case sifflet.PublicDatasetsScopeDtoTypeINCLUSION:
		return "INCLUSION", nil
	case sifflet.PublicDatasetsScopeDtoTypeEXCLUSION:
		return "EXCLUSION", nil
	default:
		return "", fmt.Errorf("unsupported value for Datasets Scope type: %s", v)
	}
}

func ToPublicDatasetsScopeDto(ctx context.Context, scopeObject types.Object) (*sifflet.PublicDatasetsScopeDto, diag.Diagnostics) {
	if scopeObject.IsNull() || scopeObject.IsUnknown() {
		return nil, diag.Diagnostics{}
	}
	var scopeDto sifflet.PublicDatasetsScopeDto
	var scopeModel DatasetsScopeModel
	diags := scopeObject.As(ctx, &scopeModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	var datasets []string
	diags = scopeModel.Datasets.ElementsAs(ctx, &datasets, false)
	if diags.HasError() {
		return nil, diags
	}

	scopeType, err := parseDatasetsScopeType(scopeModel.Type.ValueString())
	if err != nil {
		return nil, tfutils.ErrToDiags("Cannot create Datasets Scope Dto", err)
	}
	scopeDto.Datasets = &datasets
	scopeDto.Type = scopeType

	return &scopeDto, diag.Diagnostics{}
}

func FromPublicDatasetsScopeDto(ctx context.Context, scopeDto *sifflet.PublicDatasetsScopeDto) (types.Object, diag.Diagnostics) {
	if scopeDto == nil {
		return types.ObjectNull(DatasetsScopeTypeAttributes.AttrTypes), diag.Diagnostics{}
	}
	var scopeModel DatasetsScopeModel
	scopeTypeString, err := datasetsScopeTypeToString(scopeDto.Type)
	if err != nil {
		return types.ObjectNull(DatasetsScopeTypeAttributes.AttrTypes), tfutils.ErrToDiags("Cannot read Datasets Scope Dto", err)
	}
	scopeDatasets, diags := types.ListValueFrom(ctx, types.StringType, scopeDto.Datasets)
	if diags.HasError() {
		return types.ObjectNull(DatasetsScopeTypeAttributes.AttrTypes), diags
	}

	scopeModel.Type = types.StringValue(scopeTypeString)
	scopeModel.Datasets = scopeDatasets
	scopeObject, diags := types.ObjectValueFrom(ctx, DatasetsScopeTypeAttributes.AttrTypes, scopeModel)
	if diags.HasError() {
		return types.ObjectNull(DatasetsScopeTypeAttributes.AttrTypes), diags
	}
	return scopeObject, diag.Diagnostics{}
}
