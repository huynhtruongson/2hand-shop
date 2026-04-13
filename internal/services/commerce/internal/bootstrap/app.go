package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger/zerolog"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	mqmanager "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/manager"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/infrastructure/postgres"
	mqrabbitmq "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/infrastructure/rabbitmq"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/infrastructure/stripe"
	appHttp "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/transports/http"
)

// App holds all managed dependencies for the commerce service.
type App struct {
	cfg             config.Config
	db              postgressqlx.DB
	logger          logger.Logger
	httpServer      *appHttp.HttpServer
	rabbitmqManager mqmanager.Manager
}

// NewApp constructs the commerce service, wiring every dependency.
// Call Run() on the returned App to start the service.
func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	appLogger := zerolog.NewZeroLogger(zerolog.Config{
		Level:       cfg.Logger.Level,
		ServiceName: cfg.App.ServiceName,
		Environment: cfg.App.Environment,
	})

	db, err := postgressqlx.NewDB(postgressqlx.Config{
		Host:          cfg.Postgres.Host,
		Port:          cfg.Postgres.Port,
		User:          cfg.Postgres.User,
		Password:      cfg.Postgres.Password,
		Name:          cfg.Postgres.DBName,
		SSLMode:       cfg.Postgres.SSLMode,
		EnableLogging: true,
	}, appLogger)
	if err != nil {
		appLogger.Fatal("failed to connect postgres", "error", err)
	}

	stripe.Init(cfg.Stripe)

	// RabbitMQ is required — fail fast if unavailable.
	mgr, err := mqrabbitmq.NewRabbitMQManager(cfg.RabbitMQ, appLogger, nil)
	if err != nil {
		appLogger.Fatal("failed to create rabbitmq manager", "error", err)
	}

	// Infrastructure layer.
	cartRepo := postgres.NewCartRepo()
	orderRepo := postgres.NewOrderRepo()
	paymentRepo := postgres.NewPaymentRepo()
	checkoutProv := stripe.NewPaymentProvider(appLogger)

	// Application layer — wire all command and query handlers.
	app := application.Application{
		Commands: application.Commands{
			AddToCart:             command.NewAddToCartHandler(cartRepo, db),
			RemoveFromCart:        command.NewRemoveFromCartHandler(cartRepo, db),
			CreateCheckoutSession: command.NewCreateCheckoutSessionHandler(cartRepo, orderRepo, checkoutProv, db),
			CompleteCheckout:      command.NewCompleteCheckoutHandler(cartRepo, orderRepo, paymentRepo, db, mgr.Producer(), appLogger),
		},
		Queries: application.Queries{
			GetCart:            query.NewGetCartHandler(cartRepo, db),
			GetCheckoutSession: query.NewGetCheckoutSessionHandler(checkoutProv),
		},
	}

	// Transport layer.
	commerceHandler := appHttp.NewCommerceHandler(app)
	httpServer := appHttp.NewHttpServer(cfg, appLogger, commerceHandler)

	return &App{
		cfg:             cfg,
		db:              db,
		logger:          appLogger,
		httpServer:      httpServer,
		rabbitmqManager: mgr,
	}
}

// Run starts the HTTP server and RabbitMQ consumers, then blocks until a
// shutdown signal (SIGINT / SIGTERM) is received.
func (a *App) Run() {
	a.logger.Info(fmt.Sprintf("Starting %s service", a.cfg.App.ServiceName))

	// Start HTTP server in background.
	go func() {
		a.logger.Info(fmt.Sprintf("%s HTTP server listening on %s", a.cfg.App.ServiceName, a.httpServer.Addr()))
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal("http server error", "error", err)
		}
	}()

	// Start RabbitMQ consumers in background.
	go func() {
		if err := a.rabbitmqManager.Start(context.Background()); err != nil {
			a.logger.Error("rabbitmq manager start failed", "error", err)
		}
	}()

	// Wait for shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Fatal("http server forced to shutdown", "error", err)
	}

	if err := a.rabbitmqManager.Stop(); err != nil {
		a.logger.Error("rabbitmq manager stop failed", "error", err)
	}

	if err := a.db.Close(); err != nil {
		a.logger.Fatal("close db failed", "error", err)
	}

	a.logger.Info("Shutdown complete")
}
