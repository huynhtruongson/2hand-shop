package postgressqlx

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/middleware"
	"github.com/lib/pq"
	"github.com/qustavo/sqlhooks/v2"
)

// contextKey is a custom type to avoid collisions in context values.
type contextKey string

const startTimeKey contextKey = "sql_start_time"

var (
	driverOnce sync.Once
)

func WrapDriver(l logger.Logger) {
	driverOnce.Do(func() {
		sql.Register("postgres+hooks", sqlhooks.Wrap(&pq.Driver{}, &sqlHooks{logger: l}))
	})
}

// sqlHooks implements sqlhooks.Hooks backed by the shared logger.
type sqlHooks struct {
	logger logger.Logger
}

func (h *sqlHooks) Before(ctx context.Context, _ string, _ ...interface{}) (context.Context, error) {

	return context.WithValue(ctx, startTimeKey, time.Now()), nil
}

func (h *sqlHooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	start, ok := ctx.Value(startTimeKey).(time.Time)
	if !ok {
		return ctx, nil
	}
	reqID := middleware.GetRequestIDFromCtx(ctx)
	duration := time.Since(start)
	operation := extractOperation(query)

	cleanQuery := strings.ReplaceAll(strings.ReplaceAll(query, "\n", " "), "\t", " ")
	h.logger.Debug(
		"executed sql",
		"sql_query", cleanQuery,
		"duration_ms", duration.Seconds()*1000,
		"sql_operation", operation,
		"sql_success", true,
		"request_id", reqID,
	)

	return ctx, nil
}

func (h *sqlHooks) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	reqID := middleware.GetRequestIDFromCtx(ctx)

	start, _ := ctx.Value(startTimeKey).(time.Time)
	duration := time.Since(start)
	operation := extractOperation(query)

	cleanQuery := strings.ReplaceAll(strings.ReplaceAll(query, "\n", " "), "\t", " ")
	h.logger.Error(
		"sql error",
		"sql_query", cleanQuery,
		"sql_args", args,
		"duration_ms", duration.Seconds()*1000,
		"sql_operation", operation,
		"sql_success", false,
		"error", err,
		"request_id", reqID,
	)

	return err
}

func extractOperation(query string) string {
	parts := strings.Fields(query)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}
