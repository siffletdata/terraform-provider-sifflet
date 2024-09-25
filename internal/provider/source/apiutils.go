package source

import (
	"context"
	"fmt"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func handleDtoToModelError(err error, sourceType string) diag.Diagnostics {
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to read source",
				fmt.Sprintf(
					"couldn't parse parameters for source type %s",
					sourceType,
				),
			),
		}
	}
	return diag.Diagnostics{}
}

func (m ParametersModel) AsCreateSourceDto(ctx context.Context) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	if m.SourceType.IsNull() || m.SourceType.IsUnknown() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to create source",
				"The source type in the plan data is null or unknown, can't proceed. This is bug in the provider code.",
			),
		}
	}
	st := m.SourceType.ValueString()
	sourceParamsType, err := ParamsImplFromSchemaName(st)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}

	dto, diags := sourceParamsType.CreateSourceDtoFromModel(ctx, m)
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	return dto, diag.Diagnostics{}

}

func (m ParametersModel) AsUpdateSourceDto(ctx context.Context) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	if m.SourceType.IsNull() || m.SourceType.IsUnknown() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to update source",
				"The source type in the plan data is null or unknown, can't proceed. This is bug in the provider code.",
			),
		}
	}
	st := m.SourceType.ValueString()
	sourceParamsType, err := ParamsImplFromSchemaName(st)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}

	dto, diags := sourceParamsType.UpdateSourceDtoFromModel(ctx, m)
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	return dto, diag.Diagnostics{}

}
