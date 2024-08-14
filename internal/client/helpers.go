package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func HandleHttpErrorAsProblem(ctx context.Context, diagnostics *diag.Diagnostics, summary string, statusCode int, body []byte) {
	var problem Problem

	if err := json.Unmarshal(body, &problem); err != nil {
		// Could not parse the provided body as a generic Problem (maybe the API isn't returning this type
		// on error). This is not a fatal error as this function could be used with APIs that don't use
		// the Problem type on error.
		tflog.Error(ctx, fmt.Sprintf("Failed to unmarshal response body as a Problem: %s", err))
		diagnostics.AddError(
			summary,
			fmt.Sprintf("HTTP status code: %d", statusCode),
		)
	}

	diagnostics.AddError(
		summary,
		fmt.Sprintf("HTTP status code: %d. Details: %s %s", statusCode, *problem.Title, *problem.Detail),
	)

}
