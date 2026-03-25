package zerolog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/rs/zerolog"
)

type zeroLogger struct {
	zl *zerolog.Logger
}

type Config struct {
	Level       string
	ServiceName string
	Environment string
}

func NewZeroLogger(cfg Config) logger.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var w io.Writer = os.Stderr
	if cfg.Environment == "local" {
		w = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
			FormatLevel: func(i interface{}) string {
				return strings.ToUpper(fmt.Sprintf("[%s]", i))
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("| %s |", i)
			},
			FormatCaller: func(i interface{}) string {
				return filepath.Base(fmt.Sprintf("%s", i))
			}}

	}

	base := zerolog.New(w).With().Timestamp().Caller().Logger()

	if cfg.ServiceName != "" {
		base = base.With().Str("service", cfg.ServiceName).Logger()
	}
	if cfg.Environment != "" {
		base = base.With().Str("env", cfg.Environment).Logger()
	}

	return &zeroLogger{zl: &base}
}

func (l *zeroLogger) Debug(msg string, kv ...any) {
	applyFields(l.zl.Debug(), kv).Msg(msg)
}

func (l *zeroLogger) Info(msg string, kv ...any) {
	applyFields(l.zl.Info(), kv).Msg(msg)
}

func (l *zeroLogger) Warn(msg string, kv ...any) {
	applyFields(l.zl.Warn(), kv).Msg(msg)
}

func (l *zeroLogger) Error(msg string, kv ...any) {
	applyFields(l.zl.Error(), kv).Msg(msg)
}

func (l *zeroLogger) Fatal(msg string, kv ...any) {
	applyFields(l.zl.Fatal(), kv).Msg(msg)
}

func (l *zeroLogger) With(kv ...any) logger.Logger {
	ctx := l.zl.With()
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		switch v := kv[i+1].(type) {
		case string:
			ctx = ctx.Str(key, v)
		case int:
			ctx = ctx.Int(key, v)
		case bool:
			ctx = ctx.Bool(key, v)
		case float64:
			ctx = ctx.Float64(key, v)
		case float32:
			ctx = ctx.Float32(key, v)
		case error:
			ctx = ctx.AnErr(key, v)
		default:
			ctx = ctx.Interface(key, v)
		}
	}
	zl := ctx.Logger()
	return &zeroLogger{zl: &zl}
}

func applyFields(e *zerolog.Event, kv []any) *zerolog.Event {
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		switch v := kv[i+1].(type) {
		case string:
			e = e.Str(key, v)
		case error:
			e = e.AnErr(key, v)
		case int:
			e = e.Int(key, v)
		case bool:
			e = e.Bool(key, v)
		default:
			e = e.Interface(key, v)
		}
	}
	return e
}
