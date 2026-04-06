package http

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
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

func (h *CatalogHandler) SearchProductsHandler(ctx *gin.Context) {
	var req dto.SearchProductsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	result, err := h.app.Queries.SearchProducts.Handle(ctx, req.ToSearchProductsQuery())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseWithPagination(ctx, dto.ToSearchProductsDTO(result.Products), &result.Pagination)
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

func (h *CatalogHandler) CreateProductRequestHandler(ctx *gin.Context) {
	var req dto.CreateProductRequestDTO

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, caterrors.ErrUnauthorized)
		return
	}
	req.SellerID = authUser.UserID()

	result, err := h.app.Commands.CreateProductRequest.Handle(ctx, req.ToCreateProductRequestCommand())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.CreateProductRequestResponseDTO{ProductRequestID: result.ProductRequestID})
}

func (h *CatalogHandler) PublishProductHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	_, err := h.app.Commands.PublishProduct.Handle(ctx, command.PublishProductCommand{
		ProductID: reqID.ProductID,
	})
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

func (h *CatalogHandler) ListProductRequestsHandler(ctx *gin.Context) {
	var req dto.ListProductRequestsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	query := req.ToListProductRequestsQuery()

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, caterrors.ErrUnauthorized)
		return
	}
	query.User = &authUser

	result, err := h.app.Queries.ListProductRequests.Handle(ctx, query)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseWithPagination(ctx, dto.ToProductRequestsDTO(result.ProductRequests), &result.Pagination)
}

func (h *CatalogHandler) UpdateProductRequestHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	var req dto.UpdateProductRequestDTO
	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, caterrors.ErrUnauthorized)
		return
	}

	result, err := h.app.Commands.UpdateProductRequest.Handle(ctx,
		req.ToUpdateProductRequestCommand(reqID.ProductID, authUser.UserID()))
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, result)
}

func (h *CatalogHandler) DeleteProductRequestHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, caterrors.ErrUnauthorized)
		return
	}

	_, err := h.app.Commands.DeleteProductRequest.Handle(ctx, command.DeleteProductRequestCommand{
		ProductRequestID: reqID.ProductID,
		SellerID:         authUser.UserID(),
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.DeleteProductRequestResponseDTO{})
}

func (h *CatalogHandler) AcceptProductRequestHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestPathID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	result, err := h.app.Commands.AcceptProductRequest.Handle(ctx, command.AcceptProductRequestCommand{
		ProductRequestID: reqID.ProductRequestID,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.AcceptProductRequestResponseDTO{ProductID: result.ProductID})
}

func (h *CatalogHandler) RejectProductRequestHandler(ctx *gin.Context) {
	var reqID dto.ProductRequestPathID
	if err := ctx.ShouldBindUri(&reqID); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	var reqBody dto.RejectProductRequestDTO
	if err := ctx.ShouldBind(&reqBody); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	_, err := h.app.Commands.RejectProductRequest.Handle(ctx, command.RejectProductRequestCommand{
		ProductRequestID:  reqID.ProductRequestID,
		AdminRejectReason: reqBody.AdminRejectReason,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.RejectProductRequestResponseDTO{})
}
