package transport

import (
	"log/slog"
	"net/http"
	"support_chat/pkg/logger"
)

func LoggingMiddleware(baseLogger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqLogger := baseLogger.With("method", r.Method, "path", r.URL.Path)
		ctx := logger.WithContext(r.Context(), reqLogger)
		reqWithContext := r.WithContext(ctx)

		next.ServeHTTP(w, reqWithContext)
	})
}
