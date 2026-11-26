package middleware

import (
	"net/http"

	"avito/internal/config"
	"avito/internal/utils"
	"avito/pkg"
)

func BasicAuthMiddleWare(cfg *config.Config, log pkg.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user, pass, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="pprof"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if user != cfg.Login && pass != cfg.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="pprof"`)
				w.WriteHeader(http.StatusUnauthorized)
				utils.WriteError(w, http.StatusUnauthorized, "invalid login or password")
				return
			}
			next.ServeHTTP(w, r)
		})

	}
}
