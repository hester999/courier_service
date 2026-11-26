package monitoring

import "github.com/prometheus/client_golang/prometheus"

func Init() {
	prometheus.MustRegister(HttpRequestsTotal, HttpRequestDuration, GateWayRetryTotal, RateLimiterExceededTotal)
}

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	GateWayRetryTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_retries_total",
		Help: "total number of gateway retry",
	},
		nil,
	).WithLabelValues()

	RateLimiterExceededTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "total number of rate limit exceeded",
		},
		[]string{"method", "path", "status"},
	)
)
