package main

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"avito/internal/gateway/order"
	"avito/internal/infra/monitoring"
	"avito/internal/infra/ratelimiter"

	"avito/cli"
	"avito/internal/config"
	ch "avito/internal/handler/courier"
	dh "avito/internal/handler/delivery"
	sh "avito/internal/handler/service"
	orderpoller "avito/internal/infra/background/get_orders"
	statusCheck "avito/internal/infra/background/status_checker"
	cachePkg "avito/internal/infra/cache"
	"avito/internal/infra/db"
	cr "avito/internal/repo/courier"
	dr "avito/internal/repo/delivery"
	"avito/internal/router"
	cu "avito/internal/usecases/courier"
	usecaseChecker "avito/internal/usecases/courier/status_check"
	du "avito/internal/usecases/delivery"
	"avito/internal/usecases/delivery/factory"
	"avito/pkg"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/dig"
)

func main() {
	container := dig.New()

	if err := provide(container); err != nil {
		panic(err)
	}

	if err := container.Invoke(setupConfig); err != nil {
		panic(err)
	}

	if err := container.Invoke(runServer); err != nil {
		panic(err)
	}
}

func provide(c *dig.Container) error {
	// --- LOGGER ---
	if err := c.Provide(func() pkg.Logger {
		return pkg.NewLogger(0)
	}); err != nil {
		return err
	}

	// --- CONFIG & CLI ---
	if err := c.Provide(config.NewConfig); err != nil {
		return err
	}

	if err := c.Provide(cli.NewFlagsCLI); err != nil {
		return err
	}

	monitoring.Init()

	// --- DATABASE ---
	if err := c.Provide(func(cfg *config.Config) (*sqlx.DB, error) {
		return db.InitDataBaseConnection(cfg.MaxAttempts, cfg.Delay)
	}); err != nil {
		return err
	}

	// --- TRANSACTION MANAGER ---
	if err := c.Provide(func(db *sqlx.DB) (*manager.Manager, *trmsqlx.CtxGetter, error) {
		factory := trmsqlx.NewDefaultFactory(db)
		mgr := manager.Must(factory)
		getter := trmsqlx.DefaultCtxGetter
		return mgr, getter, nil
	}); err != nil {
		return err
	}

	// --- TxRunner (используется в usecase) ---
	if err := c.Provide(func(mgr *manager.Manager) du.TxRunner {
		return db.NewTRMRunner(mgr)
	}); err != nil {
		return err
	}
	// gate
	if err := c.Provide(func(cfg *config.Config, log pkg.Logger) *order.GRPCOrderGateway {
		res, err := order.NewGRPCOrderGateway(cfg.OrderServiceAddress, cfg.ConnRetryToOrderService, cfg.ConnectDelay, cfg.RetryDelay, monitoring.GateWayRetryTotal, log)
		if err != nil {
			return nil
		}
		return res
	}); err != nil {
		return err
	}

	// --- REPOSITORIES ---
	// if err := c.Provide(func(db *sqlx.DB, getter *trmsqlx.CtxGetter, log pkg.Logger) *cr.RepoCourier {
	// 	return cr.NewRepoCourier(db, getter, log)
	// }); err != nil {
	// 	return err
	// }

	if err := c.Provide(cr.NewRepoCourier); err != nil {
		return err
	}

	// if err := c.Provide(func(db *sqlx.DB, getter *trmsqlx.CtxGetter, log pkg.Logger) *dr.DeliveryRepo {
	// 	return dr.NewDeliveryRepo(db, getter, log)
	// }); err != nil {
	// 	return err
	// }
	if err := c.Provide(dr.NewDeliveryRepo); err != nil {
		return err
	}

	// --- USECASES ---
	if err := c.Provide(func(repo *cr.RepoCourier, log pkg.Logger) *cu.UsecasesCourier {
		return cu.NewUsecasesCourier(repo, log)
	}); err != nil {
		return err
	}

	// DeliveryUsecase
	if err := c.Provide(func(
		tx du.TxRunner,
		dr *dr.DeliveryRepo,
		cr *cr.RepoCourier,
		fabric *factory.DeadLineFactory,
		gateway *order.GRPCOrderGateway,

		log pkg.Logger,
	) *du.DeliveryUsecase {
		return du.NewDeliveryUsecase(tx, dr, cr, fabric, gateway, log)
	}); err != nil {
		return err
	}

	// --- HANDLERS ---
	if err := c.Provide(func(uc *cu.UsecasesCourier, log pkg.Logger) *ch.HandlerCourier {
		return ch.NewHandlerCourier(uc, log)
	}); err != nil {
		return err
	}

	// DeliveryHandler
	if err := c.Provide(func(uc *du.DeliveryUsecase, log pkg.Logger) *dh.DeliveryHandler {
		return dh.NewDeliveryHandler(uc, log)
	}); err != nil {
		return err
	}

	if err := c.Provide(sh.NewHandler); err != nil {
		return err
	}

	// --- ROUTER ---
	if err := c.Provide(router.NewRouter); err != nil {
		return err
	}
	if err := c.Provide(factory.NewDeadLineFabric); err != nil {
		return err
	}

	// --- CACHE ---
	if err := c.Provide(func(cfg *config.Config, log pkg.Logger) *cachePkg.CacheDeadlineDelivery {
		// todo cfg
		return cachePkg.NewLRUDeadlineCache(cfg.CacheTTL, cfg.BatchSize, log)
	}); err != nil {
		return err
	}

	if err := c.Provide(func(cr *cr.RepoCourier, dr *dr.DeliveryRepo, cache *cachePkg.CacheDeadlineDelivery, tx du.TxRunner, log pkg.Logger, cfg *config.Config) *usecaseChecker.CourierStatusChecker {
		return usecaseChecker.NewCourierStatusChecker(dr, cr, cache, tx, log, cfg.BatchSize)
	}); err != nil {
		return err
	}

	if err := c.Provide(func(checker *usecaseChecker.CourierStatusChecker, log pkg.Logger, cfg *config.Config) *statusCheck.Checker {
		checkInterval := cfg.PeriodCheck
		return statusCheck.NewChecker(checker, log, checkInterval)
	}); err != nil {
		return err
	}

	if err := c.Provide(func(poller *du.DeliveryUsecase, cfg *config.Config, log pkg.Logger) *orderpoller.PollingOrder {
		return orderpoller.NewPollingOrder(poller, cfg.PollingPeriod, log)
	}); err != nil {
		return err
	}

	if err := c.Provide(func(cfg *config.Config) *ratelimiter.IpRateLimiter {
		return ratelimiter.NewIpRateLimiter(cfg.IpTTL, cfg.MaxTokens, cfg.RefillRate)
	}); err != nil {
		return err
	}

	return nil
}

func setupConfig(cfg *config.Config, cliFlag *cli.FlagsCLI, logger pkg.Logger) error {
	cfg.SetCFGFromEnvFile()

	if err := cliFlag.SetFlags(); err != nil {
		logger.Error("Error setting flags")
		return err
	}

	cfg.LoadConfig()
	return nil
}

func runServer(
	r *mux.Router,
	cfg *config.Config,
	db *sqlx.DB,
	logger pkg.Logger,
	checker *statusCheck.Checker,
	poller *orderpoller.PollingOrder,
) error {
	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
	profRouter := router.NewProfRouter(cfg, logger)
	prof := &http.Server{
		Addr:    ":" + cfg.ProfilePort, // cfg
		Handler: profRouter,
	}

	go checker.Run(ctx)
	go poller.Start(ctx)
	errChanPoller := poller.GetErrChan()

	go func(ch <-chan error) { //  нужно чтобы не забивать канал, так как от юзкейса идут ошибки
		for range ch {
		}
	}(errChanPoller)

	go func() {
		logger.Info("Server started on %s", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error: %v", err)
			stop()
		}
	}()

	if cfg.EnableProfiling {
		go func() {
			logger.Info("Server profiler started on %s", ":6000")
			if err := prof.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				logger.Error("Server error: %v", err)
				stop()
			}
		}()
	}
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed: %v", err)
	}
	if err := prof.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed: %v", err)
	}

	logger.Info("Service stopped.")
	return nil
}
