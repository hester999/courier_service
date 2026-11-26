package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"avito/internal/infra/monitoring"
	"avito/pkg"
)

type httpWriter struct {
	http.ResponseWriter
	code int
}

func (w *httpWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func PrometheusMiddleware(log pkg.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			path := metricPath(r)
			start := time.Now()
			writer := &httpWriter{
				ResponseWriter: w,
				code:           http.StatusOK,
			}
			next.ServeHTTP(writer, r)
			timer := time.Since(start).Seconds()

			monitoring.HttpRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(writer.code)).Inc()
			monitoring.HttpRequestDuration.WithLabelValues(path).Observe(timer)
			if writer.code == http.StatusTooManyRequests {
				monitoring.RateLimiterExceededTotal.WithLabelValues(r.Method, path, strconv.Itoa(writer.code)).Inc()
			}
			log.Monitoring("method=%s path=%s status=%d duration=%f", r.Method, path, writer.code, timer)
		})
	}

}

func metricPath(r *http.Request) string {
	var reMuxVars = regexp.MustCompile(`/\{[^/]+\}`)
	if route := mux.CurrentRoute(r); route != nil {
		if tmpl, err := route.GetPathTemplate(); err == nil && tmpl != "" {
			p := reMuxVars.ReplaceAllString(tmpl, "")
			if p == "" {
				return "/"
			}
			return p
		}
	}
	return r.URL.Path
}
