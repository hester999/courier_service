package router

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	cfg "avito/internal/config"
	"avito/internal/handler/courier"
	"avito/internal/handler/delivery"
	"avito/internal/handler/service"
	"avito/internal/infra/ratelimiter"
	"avito/internal/middleware"
	"avito/pkg"

	"github.com/gorilla/mux"
)

func NewRouter(serviceHandler *service.Handler, courierHandler *courier.HandlerCourier, deliveryHandler *delivery.DeliveryHandler, ipLimiter *ratelimiter.IpRateLimiter, log pkg.Logger) *mux.Router {
	router := mux.NewRouter()

	router.Handle("/metrics", promhttp.Handler())
	router.Use(middleware.PrometheusMiddleware(log))
	router.Use(middleware.RateLimiterMiddleWare(ipLimiter, log))
	router.HandleFunc("/healthcheck", serviceHandler.HealthCheck).Methods("HEAD")
	router.HandleFunc("/ping", serviceHandler.Ping).Methods("GET")

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/courier/{id}", courierHandler.GetCourier).Methods("GET") //  теперь в единственном числе, кроме списка. Для использования готовой postman коллекции, придется изменить ручку или в postman, или тут
	api.HandleFunc("/couriers", courierHandler.GetCouriers).Methods("GET")
	api.HandleFunc("/courier", courierHandler.CreateCourier).Methods("POST")
	api.HandleFunc("/courier", courierHandler.UpdateCourier).Methods("PUT")

	api.HandleFunc("/delivery/assign", deliveryHandler.AssignCourierOnOrder).Methods("POST")
	api.HandleFunc("/delivery/unassign", deliveryHandler.UnassignCourierOnOrder).Methods("POST")

	return router
}

func NewProfRouter(config *cfg.Config, log pkg.Logger) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.BasicAuthMiddleWare(config, log))
	r.HandleFunc("/debug/pprof/", pprof.Index).Methods(http.MethodGet)

	r.HandleFunc("/debug/pprof/profile", pprof.Profile).Methods(http.MethodGet)

	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Methods(http.MethodGet)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace).Methods(http.MethodGet)

	r.Handle("/debug/pprof/heap", pprof.Handler("heap")).Methods(http.MethodGet)
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine")).Methods(http.MethodGet)
	r.Handle("/debug/pprof/allocs", pprof.Handler("allocs")).Methods(http.MethodGet)
	r.Handle("/debug/pprof/block", pprof.Handler("block")).Methods(http.MethodGet)
	r.Handle("/debug/pprof/mutex", pprof.Handler("mutex")).Methods(http.MethodGet)
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate")).Methods(http.MethodGet)

	return r
}
