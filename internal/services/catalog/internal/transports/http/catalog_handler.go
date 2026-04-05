package http

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
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

func (h *CatalogHandler) ListProductHandler(ctx *gin.Context) {
	var req dto.ListProductRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	query := req.ToListProductQuery()
	authUser, ok := auth.UserFromCtx(ctx)
	if ok {
		query.User = &authUser
	}
	result, err := h.app.Queries.ListProduct.Handle(ctx, query)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseWithPagination(ctx, dto.ToProductsDTO(result.Products), &result.Pagination)
}

func (h *CatalogHandler) GetProductHandler(ctx *gin.Context) {
	var req dto.GetProductRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	product, err := h.app.Queries.GetProduct.Handle(ctx, req.ToGetProductQuery())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.ToProductDTO(*product))
}

func (h *CatalogHandler) UpdateProductHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	var req dto.UpdateProductRequest
	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	req.ProductID = reqID.ProductID

	_, err := h.app.Commands.UpdateProduct.Handle(ctx, req.ToUpdateProductCommand())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, nil)
}

func (h *CatalogHandler) DeleteProductHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	_, err := h.app.Commands.DeleteProduct.Handle(ctx, command.DeleteProductCommand{ProductID: reqID.ProductID})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, nil)
}
