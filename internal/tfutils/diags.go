// package tfutils contains small utility functions for working with the Terraform provider framework.
// These functions implements common helpers that are useful across the provider.
package tfutils

import "github.com/hashicorp/terraform-plugin-framework/diag"

// ErrToDiags converts an error to a diag.Diagnostics.
// The provided summary is used as the diagnostic summary, while the error string
// is use as the dianostic details.
func ErrToDiags(summary string, err error) diag.Diagnostics {
	if err == nil {
		return diag.Diagnostics{}
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			summary,
			err.Error(),
		),
	}
}
