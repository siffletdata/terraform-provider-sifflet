// Package tfhttp provides a [http.Client] suitable for use in this Terraform provider.
package tfhttp

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// newTerraformHttpClient creates a new http.Client with additional configuration for use in the Terraform provider.
// It can log HTT¨P requests and responses.
func NewTerraformHttpClient() *http.Client {
	rt := loggingRoundTripper{}
	return &http.Client{
		Transport: rt,
	}
}

type loggingRoundTripper struct{}

func (t loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	ctx := req.Context()
	ctx = tflog.SetField(ctx, "http.request.url", req.URL.String())
	ctx = tflog.SetField(ctx, "http.request.method", req.Method)
	tflog.Trace(ctx, "HTTP Request")

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	ctx = tflog.SetField(ctx, "http.response.status", resp.Status)
	tflog.Debug(ctx, "HTTP Response")

	return resp, err
}
