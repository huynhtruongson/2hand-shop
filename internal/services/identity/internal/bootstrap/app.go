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
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	mqmanager "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/manager"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/auth"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/eventhandler"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/infrastructure/keycloak"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/infrastructure/postgres"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/infrastructure/rabbitmq"
	appHttp "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/transports/http"
)

type App struct {
	cfg             config.Config
	db              postgressqlx.DB
	logger          logger.Logger
	httpServer      *appHttp.HttpServer
	rabbitmqManager mqmanager.Manager
}

func NewApp() *App {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := zerolog.NewZeroLogger(zerolog.Config{
		Level:       config.Logger.Level,
		ServiceName: config.App.ServiceName,
		Environment: config.App.Environment,
	})

	db, err := postgressqlx.NewDB(postgressqlx.Config{
		Host:          config.Postgres.Host,
		Port:          config.Postgres.Port,
		User:          config.Postgres.User,
		Password:      config.Postgres.Password,
		Name:          config.Postgres.DBName,
		SSLMode:       config.Postgres.SSLMode,
		EnableLogging: true,
		// MaxOpenConns:    25,
		// MaxIdleConns:    2,
		// ConnMaxLifetime: 1 * time.Minute,
		// ConnMaxIdleTime: 1 * time.Minute,
	}, logger)
	if err != nil {
		logger.Fatal("failed to connect postgres", "error", err)
	}

	dispatcher := dispatcher.NewEventDispatcher(logger, nil)
	var mqMgr mqmanager.Manager
	mqMgr, err = rabbitmq.NewRabbitMQManager(config.RabbitMQ, logger, dispatcher)
	if err != nil {
		logger.Fatal("failed to connect rabbitmq, running without message broker", "error", err)
	}

	userRepo := postgres.NewUserRepo()
	authHandler := appHttp.NewAuthHandler(auth.NewAuthService(logger, config.Keycloak))

	keycloakClient := keycloak.NewKeycloakClient(config.Keycloak)

	app := application.Application{
		Commands: application.Commands{
			UpdateProfile: command.NewUpdateProfileHandler(db, userRepo),
		},
		Queries: application.Queries{
			Profile: query.NewProfileHandler(db, userRepo),
		},
		EventHandlers: application.EventHandlers{
			OnKeycloakUserRegistration: eventhandler.NewOnKeycloakUserRegistrationHandler(logger, db, keycloakClient, userRepo),
		},
	}

	userHandler := appHttp.NewUserHandler(app)

	rabbitmq.BuildEventDispatcher(dispatcher, app.EventHandlers)

	httpServer := appHttp.NewHttpServer(config, logger, authHandler, userHandler)

	return &App{
		cfg:             config,
		db:              db,
		logger:          logger,
		httpServer:      httpServer,
		rabbitmqManager: mqMgr,
	}
}

func (a *App) Run() {
	a.logger.Info(fmt.Sprintf("Starting %s service", a.cfg.App.ServiceName))

	go func() {
		a.logger.Info(fmt.Sprintf("%s service is running on %s", a.cfg.App.ServiceName, a.httpServer.Addr()))
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal("server error", "error", err)
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
	// Graceful shutdown
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

	a.logger.Info("Shutting down server gracefully")

}
