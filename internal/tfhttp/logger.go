package tfhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// httpLogger implements the retryablehttp.LeveledLogger interface used to log HTTP requests and responses.
type httpLogger struct {
	ctx context.Context
}

var _ retryablehttp.LeveledLogger = httpLogger{}

func (l httpLogger) Error(msg string, keysAndValues ...any) {
	tflog.Error(l.ctx, msg, l.additionalFields(keysAndValues))
}

func (l httpLogger) Info(msg string, keysAndValues ...any) {
	tflog.Info(l.ctx, msg, l.additionalFields(keysAndValues))
}

func (l httpLogger) Debug(msg string, keysAndValues ...any) {
	tflog.Debug(l.ctx, msg, l.additionalFields(keysAndValues))
}

func (l httpLogger) Warn(msg string, keysAndValues ...any) {
	tflog.Warn(l.ctx, msg, l.additionalFields(keysAndValues))
}

func (l httpLogger) additionalFields(keysAndValues []any) map[string]any {
	additionalFields := make(map[string]any, len(keysAndValues))

	for i := 0; i+1 < len(keysAndValues); i += 2 {
		additionalFields[fmt.Sprint(keysAndValues[i])] = keysAndValues[i+1]
	}

	return additionalFields
}

func responseLogHook(_ retryablehttp.Logger, resp *http.Response) {
	ctx := resp.Request.Context()
	// We ignore the provided Logger, as its interface doesn't support
	// multiple log levels, and thus doesn't play nice with the TF_LOG
	// environment variable.
	tflog.Debug(ctx, "HTTP response")

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		tflog.Error(ctx, "Failed to dump response for logging", map[string]any{"error": err})
		return
	}
	respLog := fmt.Sprintf("%q", respDump)
	tflog.Trace(ctx, "HTTP response details", map[string]any{"http.response.dump": respLog})
}

func requestLogHook(_ retryablehttp.Logger, req *http.Request, retryCount int) {
	// The retryablehttp library already logs a line at the DEBUG level
	// on each request, so there's no need to additionally do that here.
	// Only log the request details at the TRACE level.
	ctx := req.Context()
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		tflog.Error(ctx, "Failed to dump request for logging", map[string]any{"error": err})
		return
	}
	reqLog := fmt.Sprintf("%q", reqDump)
	tflog.Trace(ctx, "HTTP request details", map[string]any{"http.request.dump": reqLog, "retry_count": retryCount})
}
