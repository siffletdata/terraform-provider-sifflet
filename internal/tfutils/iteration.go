package tfutils

import "github.com/hashicorp/terraform-plugin-framework/diag"

// MapWithDiagnostics applies a function that can return diagnostics to each element of a collection, and return the results and accumulated diagnostics.
func MapWithDiagnostics[T any, R any](collection []T, f func(T) (R, diag.Diagnostics)) ([]R, diag.Diagnostics) {
	result := make([]R, len(collection))
	diags := diag.Diagnostics{}
	for i, item := range collection {
		var ds diag.Diagnostics
		result[i], ds = f(item)
		diags.Append(ds...)
	}
	return result, diags
}
