package http

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/LukaGiorgadze/gonull/v2"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/middleware"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/config"
)

type HttpServer struct {
	logger logger.Logger
	cfg    config.Config
	srv    *http.Server
}

func NewHttpServer(cfg config.Config, logger logger.Logger, catalogHandler *CatalogHandler) *HttpServer {
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(nullableTypeFunc,
			gonull.Nullable[string]{},
			gonull.Nullable[int]{},
			gonull.Nullable[int32]{},
			gonull.Nullable[int64]{},
			gonull.Nullable[float32]{},
			gonull.Nullable[float64]{},
			gonull.Nullable[bool]{},
			gonull.Nullable[customtypes.Price]{},
			gonull.Nullable[customtypes.Attachment]{},
			gonull.Nullable[customtypes.Attachments]{},
		)
	}

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
	}, logger: logger}

	httpServer.registerCatalogRoutes(router, catalogHandler)

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

func (sv *HttpServer) registerCatalogRoutes(r *gin.Engine, catalogHandler *CatalogHandler) {
	// public routes
	r.GET("/products", catalogHandler.ListProductHandler)
	r.GET("/products/:product_id", catalogHandler.GetProductHandler)

	// private routes
	authMiddleware := auth.CognitoAuth(auth.CognitoConfig{
		Region:     sv.cfg.Cognito.Region,
		UserPoolID: sv.cfg.Cognito.UserPoolID,
		ClientID:   sv.cfg.Cognito.ClientID,
		TokenUse:   "access",
	})

	authRoutes := r.Group("", authMiddleware)
	authRoutes.GET("/product-requests", catalogHandler.ListProductRequestsHandler)

	// admin
	adminRoutes := r.Group("", authMiddleware, auth.RequireRole("admin"))
	adminRoutes.POST("/products", catalogHandler.CreateProductHandler)
	adminRoutes.PUT("/products/:product_id", catalogHandler.UpdateProductHandler)
	adminRoutes.DELETE("/products/:product_id", catalogHandler.DeleteProductHandler)
	adminRoutes.POST("/product-requests/:product_request_id/accept", catalogHandler.AcceptProductRequestHandler)
	adminRoutes.POST("/product-requests/:product_request_id/reject", catalogHandler.RejectProductRequestHandler)

	// client
	sellerRoutes := r.Group("", authMiddleware, auth.RequireRole("client"))
	sellerRoutes.POST("/product-requests", catalogHandler.CreateProductRequestHandler)
	sellerRoutes.PUT("/product-requests/:product_request_id", catalogHandler.UpdateProductRequestHandler)
	sellerRoutes.DELETE("/product-requests/:product_request_id", catalogHandler.DeleteProductRequestHandler)
}
func nullableTypeFunc(field reflect.Value) interface{} {
	presentField := field.FieldByName("Present")
	validField := field.FieldByName("Valid")

	// Field was not in the JSON body at all → skip validation
	if !presentField.IsValid() || !presentField.Bool() {
		return nil
	}

	// Field was explicitly null → skip value validation (omitempty)
	if !validField.IsValid() || !validField.Bool() {
		return nil
	}

	// Field has a real value → let validator inspect it
	return field.FieldByName("Val").Interface()
}
