// package model contains interfaces that should be used consistently across the provider when defining models.
// A model is a Go struct that represents some Terraform state (or plan) for a resource.
//
// Such structs will always use `tfsdk` annotations on their fields so that they can be converted to/from Terraform types.
//
// # Resource implementation with models
//
// Most of the provider logic is about converting between Terraform state and the API data model. By implementing
// these interfaces, we ensure that the provider is consistent in how it handles these conversions.
//
// Models can be nested e.g a model may contain references to other models. Inner models implement different interfaces
// than top-level models.
//
// [FullModel] should be implemented by models that implement the full top-level state for a Terraform resource
// [InnerModel] should be implemented by models that are used as part of a top-level model, e.g. a nested object in a resource.
//
// When these interfaces are implemented, the actual implementation of the resource always follow the same pattern:
//  1. load the Terraform state/plan into a model (convert Terraform types to Go types), using the Terraform plugin framework utilities
//  2. convert the model to the API data model (convert the model to a DTO), using the model's interface methods ([ToUpdateDto], [ToCreateDto])
//  3. call the API with the DTO
//  4. convert the API response to a model (convert the DTO to the model), using the model's interface methods ([FromDto])
//  5. convert the model to Terraform types, using the Terraform plugin framework utilities
//
// There's nothing that enforces this pattern (or the use of model interfaces) in resource implementations.
// You can deviate from it when the API has edge cases.
//
// # Model example
//
//	struct exampleModel {
//		  StringField types.String `tfsdk:"string_field"`
//		  NestedField types.List `tfsdk:"my_field"`
//	}
//
//	var (
//		// Ensure exampleModel implements the FullModel and ModelWithId interfaces
//		_ model.FullModel[ExampleDto, ExampleCreateDto, ExampleUpdateDto] = &exampleModel{}
//		_ model.ModelWithId[uuid.UUID] = &ExampleModel{}
//	 )
//
//	// ...
//
// Models should define helper functions to easily access nested models:
//
//	func (m exampleModel) getNestedField(ctx context.Context) ([]nestedModel, diag.Diagnostics) {
//		nested := make([]nestedModel, 0, len(m.NestedField.Elements()))
//		diags := m.NestedField.ElementsAs(ctx, &nested, false)
//		return nested, diags
//	}
package model

import (
	"context"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ReadableModel is a model that can be created from a DTO returned by the API.
type ReadableModel[T any] interface {
	// FromDto updates the model with the data from the DTO.
	FromDto(context.Context, T) diag.Diagnostics
}

// UpdatableModel is a model that can be sent to an API (after conversion to a DTO) to updated the modelized resource.
type UpdatableModel[T any] interface {
	ToUpdateDto(context.Context) (T, diag.Diagnostics)
}

// CreatableModel is a model that can be send to an API (after convertionon to a DTO) to create the modelized resource.
type CreatableModel[T any] interface {
	ToCreateDto(context.Context) (T, diag.Diagnostics)
}

// FullModel is a model that can be used to perform both read, create and update operations in the API.
type FullModel[R any, C any, U any] interface {
	ReadableModel[R]
	CreatableModel[C]
	UpdatableModel[U]
}

// InnerModel is a model that has the same DTO representation when used in read, create and update operations.
// Use this interface if the implementations of UpdatableModel and CreatableModel are the same, and ReadableModel
// uses the same type parameter as these two. This is typically used for nested models (models that are not the top-level
// Terraform state of a resource).
type InnerModel[D any] interface {
	ToDto(context.Context) (D, diag.Diagnostics)
	FromDto(context.Context, D) diag.Diagnostics

	AttributeTypes() map[string]attr.Type
}

// NewModelListFromDto is a helper function that creates a Terraform list value from a list of DTOs.
// In the function signature, D is the type of the DTO, and InnerModel[D] is the type of the model that can be created from the DTO.
// modelCtor is a function that returns a new instance of the model type that matches the DTO type.
func NewModelListFromDto[D any](ctx context.Context, dtos []D, modelCtor func() InnerModel[D]) (basetypes.ListValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// zero value used only to call AttributeTypes
	attrType := types.ObjectType{AttrTypes: modelCtor().AttributeTypes()}

	modelList, ds := tfutils.MapWithDiagnostics(dtos, func(dto D) (InnerModel[D], diag.Diagnostics) {
		model := modelCtor()
		diags := model.FromDto(ctx, dto)
		return model, diags
	})
	diags.Append(ds...)

	result, ds := types.ListValueFrom(ctx, attrType, modelList)
	diags.Append(ds...)
	return result, diags
}

// ModelWithId is a model which can be identified by a unique ID in the API.
type ModelWithId[T any] interface {
	ModelId() (T, diag.Diagnostics)
}
