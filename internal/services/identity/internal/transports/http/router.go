package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/middleware"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
)

type HttpServer struct {
	logger logger.Logger
	cfg    config.Config
	srv    *http.Server
}

func NewHttpServer(cfg config.Config, logger logger.Logger, authHandler *AuthHandler, userHandler *UserHandler) *HttpServer {
	router := gin.Default()
	router.Use(gin.Recovery(), middleware.GinRequestID(), middleware.GinLogger(middleware.LogConfig{
		Logger:          logger,
		LogRequestBody:  true,
		LogResponseBody: true,
		SkipPaths:       []string{"health"},
	}))

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := HttpServer{cfg: cfg, srv: &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GinHttp.Port),
		Handler: router,
		// ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		// WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		// IdleTimeout:  60 * time.Second,
	}, logger: logger}

	httpServer.registerAuthRoutes(router, authHandler)
	httpServer.registerUserRoutes(router, userHandler)

	return &httpServer
}

func (s *HttpServer) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *HttpServer) Addr() string {
	return s.srv.Addr
}

func (sv *HttpServer) registerAuthRoutes(r *gin.Engine, authHandler *AuthHandler) {
	r.POST("/signup", authHandler.SignUpHandler)
	r.POST("/signin", authHandler.SignInHandler)
	r.POST("/confirm-account", authHandler.ConfirmAccountHandler)
}

func (sv *HttpServer) registerUserRoutes(r *gin.Engine, userHandler *UserHandler) {
	r.Use(middleware.CognitoAuth(middleware.CognitoConfig{
		Region:     sv.cfg.Cognito.Region,
		UserPoolID: sv.cfg.Cognito.UserPoolID,
		ClientID:   sv.cfg.Cognito.ClientID,
		TokenUse:   "access",
	}))

	r.GET("/users/profile", userHandler.GetProfileHandler)
	r.PUT("/users/profile", userHandler.UpdateProfileHandler)
}
