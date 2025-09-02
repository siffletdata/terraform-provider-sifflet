// Package tfhttp provides a [http.Client] suitable for use in this Terraform provider.
package tfhttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const applicationName = "terraform-provider-sifflet"

// newTerraformHttpClient creates a new http.Client with additional configuration for use in the Terraform provider.
// It can log HTTP requests and responses.
func NewTerraformHttpClient(ctx context.Context, tfVersion string, providerVersion string) *http.Client {
	c := retryablehttp.NewClient()
	rt := retryablehttp.RoundTripper{Client: c}
	c.Logger = httpLogger{ctx: ctx}
	c.RequestLogHook = requestLogHook
	c.ResponseLogHook = responseLogHook

	transport := contentTypeValidatorRoundTripper{
		next: headersRoundTripper{
			headers: map[string]string{
				"User-Agent":         userAgent(tfVersion, providerVersion),
				"X-Application-Name": applicationName,
			},
			next: &rt,
		},
		contentTypeIncludes: "json",
	}
	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}
}

func userAgent(tfVersion string, providerVersion string) string {
	headerValue := fmt.Sprintf("Terraform/%s (+https://www.terraform.io) %s/%s", tfVersion, applicationName, providerVersion)
	if u := os.Getenv("TF_APPEND_USER_AGENT"); u != "" {
		headerValue = fmt.Sprintf("%s %s", headerValue, u)
	}
	return headerValue
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
