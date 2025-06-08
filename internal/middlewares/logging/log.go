package logging

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/arjunksofficial/tyk-task/internal/metrics"
)

// A wrapper to capture response status
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK, &bytes.Buffer{}}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b) // capture response body
	return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Serve request
		next.ServeHTTP(rw, r)

		// Log response
		duration := time.Since(start)
		// log as json format in single line
		log.Printf("{\"method\":\"%s\", \"path\":\"%s\", \"status\":%d, \"duration\":\"%s\"}",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)
		defer func() {
			duration := time.Since(start).Seconds()
			metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
			metrics.RequestDuration.WithLabelValues(r.URL.Path).Observe(duration)
		}()
	})
}
