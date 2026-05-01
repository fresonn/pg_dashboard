package logger

import (
	"context"
	"dashboard/api/internal/ctxkeys"
	"log/slog"
)

type handlerMiddleware struct {
	next slog.Handler
}

func withMiddleware(next slog.Handler) *handlerMiddleware {
	return &handlerMiddleware{next: next}
}

func (h *handlerMiddleware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

func (h *handlerMiddleware) Handle(ctx context.Context, rec slog.Record) error {

	if requestID, ok := ctxkeys.RequestID(ctx); ok {
		if requestID != "" {
			rec.AddAttrs(slog.String("request_id", requestID))
		}
	}

	return h.next.Handle(ctx, rec)
}

func (h *handlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handlerMiddleware{next: h.next.WithAttrs(attrs)}
}

func (h *handlerMiddleware) WithGroup(name string) slog.Handler {
	return &handlerMiddleware{next: h.next.WithGroup(name)}
}
