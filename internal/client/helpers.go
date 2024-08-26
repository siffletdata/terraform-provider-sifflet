package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// HandleHttpErrorAsProblem logs HTTP response bodies that have the APIProblemSchema type.
// It won't fail on other types of errors, but will log less details. Avoid using this function on API routes
// that don't return the ApiProblemSchema type on error.
func HandleHttpErrorAsProblem(ctx context.Context, diagnostics *diag.Diagnostics, summary string, httpStatusCode int, httpResponseBody []byte) {
	var problem ApiProblemSchema

	if err := json.Unmarshal(httpResponseBody, &problem); err != nil {
		// Could not parse the provided body as a generic APIProblemSchema (maybe the API isn't returning this type
		// on error). This is not a fatal error as this function could be used with APIs that don't use
		// the Problem type on error.
		tflog.Warn(ctx, fmt.Sprintf("Failed to unmarshal response body as a Problem: %s", err))
		diagnostics.AddError(
			summary,
			fmt.Sprintf("HTTP status code: %d", httpStatusCode),
		)
		return
	}

	title := ""
	if problem.Title != nil {
		title = *problem.Title
	}

	detail := ""
	if problem.Detail != nil {
		title = *problem.Detail
	}

	diagnostics.AddError(
		summary,
		fmt.Sprintf("HTTP status code: %d. Details: %s %s", httpStatusCode, title, detail),
	)

}
