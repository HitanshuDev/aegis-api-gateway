package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLogRedactsSensitiveData(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	upstream := "localhost:9000"

	handler := RequestID(
		AccessLog(
			logger,
			upstream,
			http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/users?token=query-secret",
		strings.NewReader("sensitive-body"),
	)

	req.Header.Set("Authorization", "Bearer super-secret")
	req.Header.Set("X-API-Key", "api-secret")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()

	forbidden := []string{
		"super-secret",
		"api-secret",
		"sensitive-body",
		"query-secret",
	}

	for _, value := range forbidden {
		if strings.Contains(output, value) {
			t.Errorf("sensitive value leaked into logs: %s", value)
		}
	}

	expected := []string{
		`"msg":"http_access"`,
		`"method":"POST"`,
		`"route":"/users"`,
		`"status":200`,
		`"upstream":"localhost:9000"`,
		`"request_id"`,
		`"duration_ms"`,
	}

	for _, field := range expected {
		if !strings.Contains(output, field) {
			t.Errorf("expected log field missing: %s", field)
		}
	}

	if rec.Header().Get(RequestIDHeader) == "" {
		t.Error("response missing request ID")
	}
}
