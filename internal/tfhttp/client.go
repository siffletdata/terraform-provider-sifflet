// Package tfhttp provides a [http.Client] suitable for use in this Terraform provider.
package tfhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// newTerraformHttpClient creates a new http.Client with additional configuration for use in the Terraform provider.
// It can log HTTP requests and responses.
func NewTerraformHttpClient(tfVersion string, providerVersion string) *http.Client {
	transport := contentTypeValidatorRoundTripper{
		next: loggingRoundTripper{
			next: headersRoundTripper{
				next: http.DefaultTransport,
				headers: map[string]string{
					"User-Agent": userAgent(tfVersion, providerVersion),
				},
			},
		},
		contentTypeIncludes: "json",
	}
	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}
}

func userAgent(tfVersion string, providerVersion string) string {
	header := fmt.Sprintf("Terraform/%s (+https://www.terraform.io) terraform-provider-sifflet/%s", tfVersion, providerVersion)
	if u := os.Getenv("TF_APPEND_USER_AGENT"); u != "" {
		header = fmt.Sprintf("%s %s", header, u)
	}
	return header
}

// loggingRoundTripper is an http.RoundTripper that logs requests and responses using the Terraform plugin framework.
type loggingRoundTripper struct {
	next http.RoundTripper
}

func (t loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	ctx = tflog.SetField(ctx, "http.request.url", req.URL.String())
	ctx = tflog.SetField(ctx, "http.request.method", req.Method)

	if err := logRequest(ctx, req); err != nil {
		return nil, err
	}

	resp, err := t.next.RoundTrip(req)

	// Log the status even if err is not nil
	if resp != nil {
		ctx = tflog.SetField(ctx, "http.response.status", resp.Status)
	} else {
		ctx = tflog.SetField(ctx, "http.response.status", "nil (no valid response)")
	}

	if err := logResponse(ctx, resp); err != nil {
		return nil, err
	}

	return resp, err
}

func logResponse(ctx context.Context, resp *http.Response) error {
	tflog.Debug(ctx, "HTTP response")

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		tflog.Error(ctx, "Failed to dump response for logging", map[string]interface{}{"error": err})
		return err
	}
	respLog := fmt.Sprintf("%q", respDump)
	tflog.Trace(ctx, "HTTP response details", map[string]interface{}{"http.response.dump": respLog})
	return nil
}

func logRequest(ctx context.Context, req *http.Request) error {
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		tflog.Error(ctx, "Failed to dump request for logging", map[string]interface{}{"error": err})
		return err
	}
	reqLog := fmt.Sprintf("%q", reqDump)
	tflog.Trace(ctx, "HTTP request details", map[string]interface{}{"http.request.dump": reqLog})
	return nil
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

// headersRoundTripper is an http.RoundTripper that adds headers to the request before sending it.
type headersRoundTripper struct {
	next    http.RoundTripper
	headers map[string]string
}

func (t headersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Add(k, v)
	}
	return t.next.RoundTrip(req)
}
