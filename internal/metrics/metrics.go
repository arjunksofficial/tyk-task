package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests received",
		},
		[]string{"method", "path"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	RateLimitHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total requests blocked due to rate limiting",
		},
	)

	AuthFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of failed token validations",
		},
	)
)

func Init() {
	prometheus.MustRegister(HttpRequestsTotal, RequestDuration, RateLimitHits, AuthFailures)
}
