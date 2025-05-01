package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/singl3focus/load_balancer-go/internal/adapters/primary/http"
	"github.com/singl3focus/load_balancer-go/internal/adapters/primary/http/handlers"
	"github.com/singl3focus/load_balancer-go/internal/adapters/primary/http/middleware"
	"github.com/singl3focus/load_balancer-go/internal/adapters/secondary/storage/postgres"
	"github.com/singl3focus/load_balancer-go/internal/config"
	"github.com/singl3focus/load_balancer-go/internal/domain"
	"github.com/singl3focus/load_balancer-go/internal/service"
	"github.com/singl3focus/load_balancer-go/pkg/logger"
)

func Run() {
	// [Инициализация конфига]
	cfg := config.MustLoadCfg(true)

	// [Запуск миграций, если в конфгурации указана опция]
	if cfg.Database.Migration.Option != "" {
		err := postgres.InvokeMigrator(cfg.Database.Postgres.URL, cfg.Database.Migration.Dir, "up")
		if err != nil {
			panic(err)
		}

		return
	}

	// [Инициализация логгера]
	lg := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.Enable)

	// [Инициализация стратегии и балансировщика]
	servers := make([]*domain.Server, 0, len(cfg.App.Balancer.Servers))
	for _, server := range cfg.App.Balancer.Servers {
		servers = append(servers, domain.NewServer(server.ID, server.URL))
	}

	strategy := service.GetStrategy(cfg.App.Balancer.Strategy, servers)

	lb := handlers.NewLoadBalancer(
		lg,
		strategy,
		servers,
		handlers.HealthCheckParams{
			Interval: 5 * time.Second, // TODO: hard
		},
		handlers.WithUsingHealthCheck(),
	)

	// [Инициализация хранилища]
	storage := postgres.NewDB(cfg.Database.Postgres.URL, lg)

	// [Инициализация обработичка, миддлвейров и сервера]
	rateLimiter := service.NewLimiter(
		domain.BucketConfig{
			Capacity: int64(cfg.App.RateLimiter.DefaultCapacity),
			FillRate: cfg.App.RateLimiter.DefaultFillRate,
		},
		storage,
	)
	rateLimiter.Cleanup(1 * time.Hour)

	rateLimitMiddleware := middleware.NewRateLimitterMiddleware(rateLimiter)

	clientHandler := handlers.NewClientHandler(lg, storage)

	handler := http.NewHandler(lb, clientHandler)
	handler = rateLimitMiddleware.Handle(handler)

	server := http.NewServer(
		handler,
		cfg.HTTPServer.Port,
		cfg.HTTPServer.ReadTimeout,
		cfg.HTTPServer.ReadHeaderTimeout,
		cfg.HTTPServer.WriteTimeout,
		cfg.HTTPServer.IdleTimeout,
	)

	// [Отслеживание изменений конфига]
	go config.WatchConfig(config.DefaultConfigPath, lg, cfg, func(newCfg *config.Config) {
		lg.Info("Config has been updated", "newCfg", newCfg)

		strategy := service.GetStrategy(cfg.App.Balancer.Strategy, []*domain.Server{}) // servers must be filling by balancer
		lb.ApplyConfig(newCfg.App.Balancer, strategy)
	})

	// [Начало работы сервиса]
	lg.Info("App started")

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-shutdown:
		lg.Info("Received shutdown signal")
	case err := <-serverErrors:
		lg.Error("Server error", "(err)", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		lg.Error("Graceful shutdown failed", "(err)", err.Error())
		if err := server.Close(); err != nil {
			lg.Error("Forced shutdown failed", "(err)", err.Error())
		}
	}

	if err := lb.GracefulShutdown(ctx); err != nil {
		lg.Error("Load balancer shutdown failed", "(err)", err)
	}

	// [Остановка работы сервиса]
	lg.Info("Application stopped")
}
