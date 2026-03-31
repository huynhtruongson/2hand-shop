package http

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/transports/http/dto"
)

type CatalogHandler struct {
	app application.Application
}

func NewCatalogHandler(app application.Application) *CatalogHandler {
	return &CatalogHandler{app: app}
}

func (h *CatalogHandler) CreateProductHandler(ctx *gin.Context) {
	var req dto.CreateProductRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	result, err := h.app.Commands.CreateProduct.Handle(ctx, req.ToCreateProductCommand())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.CreateProductResponse{ProductID: result.ProductID})
}
