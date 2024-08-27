// Package tfhttp provides a [http.Client] suitable for use in this Terraform provider.
package tfhttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// newTerraformHttpClient creates a new http.Client with additional configuration for use in the Terraform provider.
// It can log HTT¨P requests and responses.
func NewTerraformHttpClient() *http.Client {
	transport := contentTypeValidatorRoundTripper{
		next: loggingRoundTripper{
			next: http.DefaultTransport,
		},
		contentTypeIncludes: "json",
	}
	return &http.Client{
		Transport: transport,
	}
}

// loggingRoundTripper is an http.RoundTripper that logs requests and responses using the Terraform plugin framework.
type loggingRoundTripper struct {
	next http.RoundTripper
}

func (t loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	ctx := req.Context()
	ctx = tflog.SetField(ctx, "http.request.url", req.URL.String())
	ctx = tflog.SetField(ctx, "http.request.method", req.Method)
	tflog.Trace(ctx, "HTTP Request")

	resp, err := t.next.RoundTrip(req)

	// Log the status even if err is not nil
	ctx = tflog.SetField(ctx, "http.response.status", resp.Status)
	tflog.Debug(ctx, "HTTP Response")

	return resp, err
}

// contentTypeValidatorRoundTripper is an http.RoundTripper that ensures that the Content-Type header of the response contains a given string.
// The generated client code doesn't fail if the server returns a different content type, but doesn't parse the response body if the content type is unexpected. To avoid nil pointer exception when reading responses, we validate the content type here.
type contentTypeValidatorRoundTripper struct {
	next                http.RoundTripper
	contentTypeIncludes string
}

func (t contentTypeValidatorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if resp.ContentLength > 0 {

		if resp.Header.Get("Content-Type") == "" {
			return resp, fmt.Errorf("missing content-type header in server response")
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), t.contentTypeIncludes) {
			return resp, fmt.Errorf(
				"unexpected content-type header in server response: expected a header containing %s, got %s",
				t.contentTypeIncludes,
				resp.Header.Get("Content-Type"))
		}

	}

	return resp, err
}
