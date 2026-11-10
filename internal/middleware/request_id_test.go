package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestRequestIDGeneration(t *testing.T) {
	handler := RequestID(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			id := r.Header.Get(RequestIDHeader)

			if id == "" {
				t.Error("request ID was not generated")
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	id := rec.Header().Get(RequestIDHeader)

	if !regexp.MustCompile(
		`^[a-f0-9]{32}$`,
	).MatchString(id) {
		t.Fatalf("invalid generated request ID: %s", id)
	}
}

func TestRequestIDPreservation(t *testing.T) {
	handler := RequestID(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	req.Header.Set(RequestIDHeader, "client-request-123")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != "client-request-123" {
		t.Fatalf("expected supplied ID, got %s", got)
	}
}

func TestInvalidRequestIDIsReplaced(t *testing.T) {
	handler := RequestID(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	req.Header.Set(RequestIDHeader, "bad\nforged-log-entry")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	got := rec.Header().Get(RequestIDHeader)

	if got == "bad\nforged-log-entry" || got == "" {
		t.Fatalf("invalid ID was not replaced: %q", got)
	}
}
