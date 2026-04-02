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
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/eventhandler"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/infrastructure/postgres"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/infrastructure/rabbitmq"
	appHttp "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/transports/http"
)

// App holds all managed dependencies for the catalog service.
type App struct {
	cfg             config.Config
	db              postgressqlx.DB
	logger          logger.Logger
	httpServer      *appHttp.HttpServer
	rabbitmqManager mqmanager.Manager
}

// NewApp constructs the catalog service, wiring every dependency.
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
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		Name:     cfg.Postgres.DBName,
		SSLMode:  cfg.Postgres.SSLMode,
	})
	if err != nil {
		appLogger.Fatal("failed to connect postgres", "error", err)
	}

	// Infrastructure adapters (PostgreSQL repositories).
	productRepo := postgres.NewProductRepo()

	// Application layer — command, query, and event handlers.
	app := application.Application{
		Commands: application.Commands{
			CreateProduct: command.NewCreateProductHandler(productRepo, db),
			UpdateProduct: command.NewUpdateProductHandler(productRepo, db),
			DeleteProduct: command.NewDeleteProductHandler(productRepo, db),
		},
		Queries: application.Queries{
			ListProduct: query.NewListProductHandler(productRepo, db),
		},
		EventHandlers: application.EventHandlers{
			OnProductCreated: eventhandler.NewOnProductCreatedHandler(),
		},
	}

	// RabbitMQ is optional for local development; log and continue if unavailable.
	var mqMgr mqmanager.Manager
	mqMgr, err = rabbitmq.NewRabbitMQManager(cfg.RabbitMQ, appLogger, app)
	if err != nil {
		appLogger.Fatal("failed to connect rabbitmq, running without message broker", "error", err)
	}

	// Transport layer.
	catalogHttpHandler := appHttp.NewCatalogHttpHandler(app)
	httpServer := appHttp.NewHttpServer(cfg, appLogger, catalogHandler)

	return &App{
		cfg:             cfg,
		db:              db,
		logger:          appLogger,
		httpServer:      httpServer,
		rabbitmqManager: mqMgr,
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
	if a.rabbitmqManager != nil {
		go func() {
			if err := a.rabbitmqManager.Start(context.Background()); err != nil {
				a.logger.Error("rabbitmq manager start failed", "error", err)
			}
		}()
	}

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

	if a.rabbitmqManager != nil {
		if err := a.rabbitmqManager.Stop(); err != nil {
			a.logger.Error("rabbitmq manager stop failed", "error", err)
		}
	}

	if err := a.db.Close(); err != nil {
		a.logger.Fatal("close db failed", "error", err)
	}

	a.logger.Info("Shutdown complete")
}
