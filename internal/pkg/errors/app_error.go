package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Kind string

const (
	KindNotFound      Kind = "not_found"     // 404
	KindConflict      Kind = "conflict"      // 409
	KindValidation    Kind = "validation"    // 400
	KindBadRequest    Kind = "bad_request"   // 400
	KindUnprocessable Kind = "unprocessable" // 422
	KindUnauthorized  Kind = "unauthorized"  // 401
	KindForbidden     Kind = "forbidden"     // 403
	KindInternal      Kind = "internal"      // 500
	KindUnavailable   Kind = "unavailable"   // 503
	KindTimeout       Kind = "timeout"       // 504
	KindRateLimit     Kind = "rate_limit"    // 429

	KindPublishFailed // event couldn't be written to outbox / sent to RabbitMQ
	KindConsumeFailed // consumer failed to process an incoming event
	KindDeadLettered
)

var kindHTTPStatus = map[Kind]int{
	KindNotFound:      http.StatusNotFound,
	KindConflict:      http.StatusConflict,
	KindValidation:    http.StatusBadRequest,
	KindBadRequest:    http.StatusBadRequest,
	KindUnprocessable: http.StatusUnprocessableEntity,
	KindUnauthorized:  http.StatusUnauthorized,
	KindForbidden:     http.StatusForbidden,
	KindInternal:      http.StatusInternalServerError,
	KindUnavailable:   http.StatusServiceUnavailable,
	KindTimeout:       http.StatusGatewayTimeout,
	KindRateLimit:     http.StatusTooManyRequests,
}

type AppError struct {
	kind     Kind              // coarse category (drives HTTP status)
	code     string            // machine-readable code, e.g. "LISTING_NOT_FOUND"
	message  string            // safe message shown to the end user
	internal string            // technical detail — logs only, never sent to user
	cause    error             // wrapped original error — logs only
	meta     map[string]any    // structured fields for log context
	details  map[string]string // field-level validation messages (user-visible)
}

func NewAppError(kind Kind, code, message string) *AppError {
	return &AppError{
		kind:    kind,
		code:    code,
		message: message,
	}
}

// clone returns a shallow copy so that builder methods on sentinels are safe.
func (e *AppError) clone() *AppError {
	cp := *e
	if e.meta != nil {
		cp.meta = make(map[string]any, len(e.meta))
		for k, v := range e.meta {
			cp.meta[k] = v
		}
	}
	if e.details != nil {
		cp.details = make(map[string]string, len(e.details))
		for k, v := range e.details {
			cp.details[k] = v
		}
	}
	return &cp
}
func (e *AppError) WithInternal(msg string) *AppError {
	c := e.clone()
	c.internal = msg
	return c
}

func (e *AppError) WithInternalf(format string, args ...any) *AppError {
	return e.WithInternal(fmt.Sprintf(format, args...))
}

func (e *AppError) WithCause(cause error) *AppError {
	c := e.clone()
	c.cause = cause
	return c
}

func (e *AppError) WithMeta(key string, value any) *AppError {
	c := e.clone()
	if c.meta == nil {
		c.meta = make(map[string]any)
	}
	c.meta[key] = value
	return c
}

func (e *AppError) WithMetas(fields map[string]any) *AppError {
	c := e.clone()
	if c.meta == nil {
		c.meta = make(map[string]any, len(fields))
	}
	for k, v := range fields {
		c.meta[k] = v
	}
	return c
}

func (e *AppError) WithDetail(key, value string) *AppError {
	c := e.clone()
	if c.details == nil {
		c.details = make(map[string]string)
	}
	c.details[key] = value
	return c
}

func (e *AppError) WithDetails(details map[string]string) *AppError {
	c := e.clone()
	c.details = details
	return c
}

// error interface
func (e *AppError) Error() string {
	if e.internal != "" && e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.internal, e.cause)
	}
	if e.internal != "" {
		return fmt.Sprintf("[%s] %s", e.code, e.internal)
	}
	if e.cause != nil {
		return fmt.Sprintf("[%s] %v", e.code, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

// Unwrap enables errors.Is / errors.As chains.
func (e *AppError) Unwrap() error { return e.cause }

func (e *AppError) Kind() Kind { return e.kind }

func (e *AppError) Code() string            { return e.code }
func (e *AppError) Details() map[string]string { return e.details }

func (e *AppError) HTTPStatus() int {
	if s, ok := kindHTTPStatus[e.kind]; ok {
		return s
	}
	return http.StatusInternalServerError
}

// UserError is the JSON-serialisable struct returned to API consumers.
// It intentionally contains NO internal details whatsoever.
type UserError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"` // field-level validation messages
}

// UserFacing returns a sanitised UserError safe for inclusion in HTTP responses.
func (e *AppError) UserFacing() UserError {
	ue := UserError{
		Code:    e.code,
		Message: e.message,
	}
	if len(e.details) > 0 {
		ue.Details = e.details
	}
	return ue
}

// // LogEntry is a structured representation intended for loggers like zap or slog.
// // It includes all internal fields and is NEVER sent to the user.
// type LogEntry struct {
// 	Code     string         `json:"code"`
// 	Kind     string         `json:"kind"`
// 	Internal string         `json:"internal,omitempty"`
// 	Cause    string         `json:"cause,omitempty"`
// 	Meta     map[string]any `json:"meta,omitempty"`
// }

// // LogEntry returns a structured log entry with full internal context.
// func (e *AppError) LogEntry() LogEntry {
// 	le := LogEntry{
// 		Code:     e.code,
// 		Kind:     string(e.kind),
// 		Internal: e.internal,
// 		Meta:     e.meta,
// 	}
// 	if e.cause != nil {
// 		le.Cause = e.cause.Error()
// 	}
// 	return le
// }

// LogFields returns a flat map suitable for loggers that accept key-value pairs
// (e.g. zap.Any, slog.Attr).
func (e *AppError) LogFields() map[string]any {
	f := map[string]any{
		"error_code": e.code,
		"error_kind": string(e.kind),
	}
	if e.internal != "" {
		f["internal"] = e.internal
	}
	if e.cause != nil {
		f["cause"] = e.cause.Error()
	}
	for k, v := range e.meta {
		f[k] = v
	}
	return f
}

// As is a typed wrapper around errors.As for *AppError.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// IsKind reports whether any error in the chain has the given Kind.
func IsKind(err error, k Kind) bool {
	if ae, ok := As(err); ok {
		return ae.kind == k
	}
	return false
}

// IsCode reports whether any error in the chain has the given code string.
func IsCode(err error, code string) bool {
	if ae, ok := As(err); ok {
		return ae.code == code
	}
	return false
}

// HTTPStatus returns the HTTP status for any error.
// Falls back to 500 for unknown or non-AppErrors.
func HTTPStatus(err error) int {
	if ae, ok := As(err); ok {
		return ae.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// ToUserFacing converts any error to a UserError safe for API responses.
// Non-AppErrors are converted to a generic internal error.
func ToUserFacing(err error) UserError {
	if ae, ok := As(err); ok {
		return ae.UserFacing()
	}
	return UserError{
		Code:    "INTERNAL_ERROR",
		Message: "an unexpected error occurred",
	}
}
