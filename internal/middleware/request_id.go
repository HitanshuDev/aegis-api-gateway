package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"
)

const RequestIDHeader = "X-Request-ID"

var validRequestID = regexp.MustCompile(
	`^[a-zA-Z0-9._-]{1,128}$`,
)

var fallbackCounter atomic.Uint64

func generateRequestID() string {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err == nil {
		return hex.EncodeToString(bytes)
	}

	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), fallbackCounter.Add(1))
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		requestID := r.Header.Get(RequestIDHeader)

		if !validRequestID.MatchString(requestID) {
			requestID = generateRequestID()
		}

		r.Header.Set(RequestIDHeader, requestID)

		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r)
	})
}
