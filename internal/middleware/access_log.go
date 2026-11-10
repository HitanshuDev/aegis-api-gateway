package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(data)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func AccessLog(
	logger *slog.Logger,
	upstream string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		wrapped := &statusWriter{
			ResponseWriter: w,
		}

		defer func() {
			duration := time.Since(start)

			if !wrapped.wroteHeader {
				wrapped.status = http.StatusOK
			}

			logger.Info(
				"http_access",
				"timestamp", time.Now().UTC().Format(time.RFC3339Nano),
				"request_id", r.Header.Get(RequestIDHeader),
				"method", r.Method,
				"route", r.URL.Path,
				"status", wrapped.status,
				"duration_ms", duration.Milliseconds(),
				"upstream", upstream,
			)
		}()

		next.ServeHTTP(wrapped, r)
	})
}
