package http

import (
	"github.com/gin-gonic/gin"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/query"
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

	claim := auth.GetClaims(ctx)

	_, err := h.app.Commands.UpdateProfile.Handle(ctx.Request.Context(), command.UpdateProfile{
		UserID:    claim.UserID(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Gender:    req.Gender,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, nil)
}

func (h *UserHandler) GetProfileHandler(ctx *gin.Context) {
	claim := auth.GetClaims(ctx)

	user, err := h.app.Queries.Profile.Handle(ctx.Request.Context(), query.Profile{
		Email: claim.UserID(),
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.ToGetProfileResponse(user))
}
