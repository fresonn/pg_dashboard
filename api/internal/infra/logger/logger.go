package logger

import (
	"context"
	"dashboard/api/internal/config"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)

	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

func New(appCfg config.AppConfig) *slog.Logger {

	var handler slog.Handler

	switch appCfg.Env.Runtime {

	case config.LiveRuntime:

		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.DateTime,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

				if a.Key == "error" {
					v, ok := a.Value.Any().(error)
					if !ok {
						return a
					}
					return tint.Err(v)
				}

				return a
			},
		})

	case config.DevRuntime:

		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

				if a.Key == "error" {
					v, ok := a.Value.Any().(error)
					if !ok {
						return a
					}
					return tint.Err(v)
				}

				return a
			},
		})

	case config.TestRuntime:

		handler = slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelError + 10,
		})
	}

	handler = withMiddleware(handler)

	log := slog.New(handler)

	slog.SetDefault(log)

	return log
}

func WithScopeLogger(logger *slog.Logger, scope string) Logger {

	return logger.With("scope_name", scope)
}
