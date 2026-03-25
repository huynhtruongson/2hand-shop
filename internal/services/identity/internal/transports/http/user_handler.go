package http

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/middleware"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/transports/http/dto"
)

type UserHandler struct {
	app application.Application
}

func NewUserHandler(app application.Application) *UserHandler {
	return &UserHandler{app: app}
}

func (h *UserHandler) UpdateProfileHandler(ctx *gin.Context) {
	var req dto.UpdateProfileRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		utils.ResponseError(ctx, errors.ErrUnauthorized)
		return
	}

	_, err := h.app.Commands.UpdateProfile.Handle(ctx, command.UpdateProfile{
		UserID: userID,
		Name:   req.Name,
		Gender: req.Gender,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, nil)
}

func (h *UserHandler) GetProfileHandler(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		utils.ResponseError(ctx, errors.ErrUnauthorized)
		return
	}

	user, err := h.app.Queries.Profile.Handle(ctx, query.Profile{
		ID: userID,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.ToGetProfileResponse(user))
}
