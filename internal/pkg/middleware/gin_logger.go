package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
)

type LogConfig struct {
	// Logger is the shared logger instance (required)
	Logger logger.Logger
	// SkipPaths will not be logged
	SkipPaths []string
	// UTC sets timestamps to UTC (default: local time)
	UTC bool
	// LogRequestBody enables logging of the request body
	LogRequestBody bool
	// LogResponseBody enables logging of the response body
	LogResponseBody bool
	// MaxBodySize is the max bytes to read for request/response bodies (default: 4096)
	MaxBodySize int64
	// SkipBodyContentTypes skips body logging for these content types
	SkipBodyContentTypes []string
	// SensitiveFields are JSON keys whose values will be masked (case-insensitive)
	SensitiveFields []string
}

var defaultSensitiveFields = []string{
	"password", "secret",
	"token", "access_token", "refresh_token", "authorization",
	"credit_card", "card_number", "cvv", "cvc",
	"ssn", "social_security",
	"api_key", "apikey",
	"private_key",
}

type bodyWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (bw *bodyWriter) Write(b []byte) (int, error) {
	bw.buf.Write(b)
	return bw.ResponseWriter.Write(b)
}

func GinLogger(cfg LogConfig) gin.HandlerFunc {
	if cfg.MaxBodySize == 0 {
		cfg.MaxBodySize = 4096
	}
	// Merge default sensitive fields with user-supplied ones (deduplicated, lowercased)
	sensitiveSet := make(map[string]struct{}, len(defaultSensitiveFields)+len(cfg.SensitiveFields))
	for _, f := range defaultSensitiveFields {
		sensitiveSet[strings.ToLower(f)] = struct{}{}
	}
	for _, f := range cfg.SensitiveFields {
		sensitiveSet[strings.ToLower(f)] = struct{}{}
	}

	skipPaths := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipPaths[p] = struct{}{}
	}

	skipContentTypes := cfg.SkipBodyContentTypes
	if len(skipContentTypes) == 0 {
		skipContentTypes = []string{"multipart/form-data", "application/octet-stream"}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// ── Capture request body ──────────────────────────────────────────────
		var reqBody string
		if cfg.LogRequestBody && c.Request.Body != nil {
			ct := c.Request.Header.Get("Content-Type")
			if !shouldSkipBody(ct, skipContentTypes) {
				bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, cfg.MaxBodySize+1))
				if err == nil {
					// Restore body so downstream handlers can still read it
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					masked := maskSensitiveFields(truncateBytes(bodyBytes, cfg.MaxBodySize), sensitiveSet)
					reqBody = masked
				}
			} else {
				reqBody = "[skipped: " + ct + "]"
			}
		}

		// ── Wrap response writer to capture body ──────────────────────────────
		var bw *bodyWriter
		if cfg.LogResponseBody {
			bw = &bodyWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
			c.Writer = bw
		}

		// ── Process request ───────────────────────────────────────────────────
		c.Next()

		// ── Skip configured paths ─────────────────────────────────────────────
		if _, ok := skipPaths[path]; ok {
			return
		}

		// ── Collect post-request fields ───────────────────────────────────────
		latency := time.Since(start)
		if cfg.UTC {
			start = start.UTC()
		}

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		fullPath := path
		if raw != "" {
			fullPath = path + "?" + maskQuerySensitive(raw, sensitiveSet)
		}
		errorMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()
		requestID := GetRequestID(c)

		l := cfg.Logger.With(
			"request_id", requestID,
			"status", statusCode,
			"latency", latency.String(),
			"client_ip", clientIP,
			"method", method,
			"path", fullPath,
			"timestamp", start.Format(time.RFC3339),
		)

		// ── Request details ───────────────────────────────────────────────────
		reqKV := []any{
			"req_headers", sanitizeHeaders(c.Request.Header, sensitiveSet),
		}
		if cfg.LogRequestBody {
			if reqBody == "" {
				reqBody = "[empty]"
			}
			reqKV = append(reqKV, "req_body", reqBody)
		}

		// ── Response details ──────────────────────────────────────────────────
		respKV := []any{
			"resp_headers", sanitizeHeaders(http.Header(c.Writer.Header()), sensitiveSet),
		}
		if cfg.LogResponseBody && bw != nil {
			respBodyStr := "[empty]"
			ct := c.Writer.Header().Get("Content-Type")
			if shouldSkipBody(ct, skipContentTypes) {
				respBodyStr = "[skipped: " + ct + "]"
			} else if bw.buf.Len() > 0 {
				b := truncateBytes(bw.buf.Bytes(), cfg.MaxBodySize)
				respBodyStr = maskSensitiveFields(b, sensitiveSet)
			}
			respKV = append(respKV, "resp_body", respBodyStr)
		}

		// ── Emit log at appropriate level ─────────────────────────────────────
		allKV := append(reqKV, respKV...)
		msg := method + " " + fullPath

		switch {
		case len(errorMsg) > 0:
			l.Error(msg, append(allKV, "error", errorMsg)...)
		case statusCode >= 500:
			l.Error(msg, allKV...)
		case statusCode >= 400:
			l.Warn(msg, allKV...)
		default:
			l.Info(msg, allKV...)
		}
	}
}

// maskSensitiveFields parses body as JSON and replaces sensitive key values with "***".
// If the body is not valid JSON it is returned as-is (no mutation).
func maskSensitiveFields(body []byte, sensitiveSet map[string]struct{}) string {
	if len(body) == 0 {
		return ""
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		// Not JSON — return the raw string unchanged
		return string(body)
	}
	maskMap(raw, sensitiveSet)
	masked, err := json.Marshal(raw)
	if err != nil {
		return string(body)
	}
	return string(masked)
}

// maskMap recursively walks a decoded JSON map and redacts sensitive keys.
func maskMap(m map[string]any, sensitiveSet map[string]struct{}) {
	for k, v := range m {
		if _, isSensitive := sensitiveSet[strings.ToLower(k)]; isSensitive {
			m[k] = "***"
			continue
		}
		switch child := v.(type) {
		case map[string]any:
			maskMap(child, sensitiveSet)
		case []any:
			maskSlice(child, sensitiveSet)
		}
	}
}

func maskSlice(s []any, sensitiveSet map[string]struct{}) {
	for _, item := range s {
		if m, ok := item.(map[string]any); ok {
			maskMap(m, sensitiveSet)
		}
	}
}

// maskQuerySensitive replaces values of sensitive query params with "***".
func maskQuerySensitive(rawQuery string, sensitiveSet map[string]struct{}) string {
	parts := strings.Split(rawQuery, "&")
	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			if _, ok := sensitiveSet[strings.ToLower(kv[0])]; ok {
				parts[i] = kv[0] + "=***"
			}
		}
	}
	return strings.Join(parts, "&")
}

// sanitizeHeaders returns a flat map of headers with sensitive values masked.
func sanitizeHeaders(h http.Header, sensitiveSet map[string]struct{}) map[string]string {
	out := make(map[string]string, len(h))
	for k, vals := range h {
		if _, ok := sensitiveSet[strings.ToLower(k)]; ok {
			out[k] = "***"
		} else {
			out[k] = strings.Join(vals, "; ")
		}
	}
	return out
}

func shouldSkipBody(contentType string, skipTypes []string) bool {
	for _, t := range skipTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// truncateBytes returns at most maxLen bytes of b.
func truncateBytes(b []byte, maxLen int64) []byte {
	if int64(len(b)) <= maxLen {
		return b
	}
	return b[:maxLen]
}
