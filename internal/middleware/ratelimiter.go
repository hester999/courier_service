package middleware

import (
	"encoding/json"
	"net"
	"net/http"

	"avito/internal/infra/ratelimiter"
	"avito/pkg"
)

func RateLimiterMiddleWare(ipRate *ratelimiter.IpRateLimiter, log pkg.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				log.Error(err.Error())
				http.Error(w, "Invalid IP", http.StatusInternalServerError)
				return
			}
			limiter := ipRate.GetLimiter(ip)
			if limiter.IsAllow() {
				next.ServeHTTP(w, r)
			} else {
				log.Error("too many requests")
				w.Header().Set("Content-Type", "Application/json; charset=utf-8")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				}{Code: http.StatusTooManyRequests, Message: http.StatusText(http.StatusTooManyRequests)})
				return
			}
		})
	}
}
