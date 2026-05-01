package http

import (
	"dashboard/api/internal/ctxkeys"
	"net/http"

	"github.com/google/uuid"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := ctxkeys.WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
