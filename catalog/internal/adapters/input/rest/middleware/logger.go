package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type ResponseWriteWrapper struct {
	http.ResponseWriter
	StatusCode int
}

func (w *ResponseWriteWrapper) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}
func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapper := &ResponseWriteWrapper{ResponseWriter: w, StatusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			level := slog.LevelInfo
			if wrapper.StatusCode >= 500 {
				level = slog.LevelError
			} else if wrapper.StatusCode >= 400 {
				level = slog.LevelWarn
			}

			attrs := []any{
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", wrapper.StatusCode),
				slog.String("duration", time.Since(start).String()),
			}

			if logger.Enabled(r.Context(), slog.LevelDebug) {
				attrs = append(attrs, slog.String("user_agent", r.UserAgent()),
					slog.Any("query_params", r.URL.Query()))
			}

			logger.Log(r.Context(), level, "request processed", attrs...)

		})
	}
}
