// Package logs level supported by Cloud Logging
// https://github.com/remko/cloudrun-slog?tab=readme-ov-file
package logs

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const (
	Error        = "error"
	Counter      = "counter"
	JSON         = "json"
	TemplateData = "template_data"
	Language     = "language"
	LanguageTag  = "language_tag"
	Length       = "length"
	Country      = "country"
	Time         = "time"
	AcceptedLang = "accepted-language"
	NumDeleted   = "number_deleted"
	Key          = "key"
)

// CloudLoggingHandler that outputs JSON understood by the structured log agent.
// See https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields
type CloudLoggingHandler struct{ handler slog.Handler }

// WithCloudTraceContext Middleware that adds the Cloud Trace ID to the context
// This is used to correlate the structured logs with the Cloud Run
// request log.
func WithCloudTraceContext(h http.Handler) http.Handler {
	projectID := os.Getenv("OBLIVIATE_PROJECT_ID")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var trace string
		traceHeader := r.Header.Get("X-Cloud-Trace-Context")
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
		}
		//nolint:staticcheck
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "trace", trace)))
	})
}

func traceFromContext(ctx context.Context) string {
	trace := ctx.Value("trace")
	if trace == nil {
		return ""
	}
	return trace.(string)
}

func NewCloudLoggingHandler(logLevel slog.Level) *CloudLoggingHandler {
	return &CloudLoggingHandler{handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			} else if a.Key == slog.LevelKey {
				a.Key = "severity"
			}
			return a
		},
	})}
}

func (h *CloudLoggingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *CloudLoggingHandler) Handle(ctx context.Context, rec slog.Record) error {
	trace := traceFromContext(ctx)
	if trace != "" {
		rec = rec.Clone()
		// Add trace ID	to the record so it is correlated with the Cloud Run request log
		// See https://cloud.google.com/trace/docs/trace-log-integration
		rec.Add("logging.googleapis.com/trace", slog.StringValue(trace))
	}
	return h.handler.Handle(ctx, rec)
}

func (h *CloudLoggingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CloudLoggingHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *CloudLoggingHandler) WithGroup(name string) slog.Handler {
	return &CloudLoggingHandler{handler: h.handler.WithGroup(name)}
}
