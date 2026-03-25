package http

import (
	"github.com/gin-gonic/gin"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/auth"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/transports/http/dto"
)

type AuthHandler struct {
	AuthService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

func (ah *AuthHandler) SignUpHandler(ctx *gin.Context) {
	var req dto.SignUpRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	result, err := ah.AuthService.SignUp(ctx, req.ToSignUpParams())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	utils.Response(ctx, dto.SignUpResp{
		UserID:     result.UserID,
		IsVerified: result.IsVerified,
	})
}

func (ah *AuthHandler) SignInHandler(ctx *gin.Context) {
	var req dto.SignInRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	result, err := ah.AuthService.SignIn(ctx, req.Email, req.Password)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	utils.Response(ctx, dto.SignInResp{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

func (ah *AuthHandler) ConfirmAccountHandler(ctx *gin.Context) {
	var req dto.ConfirmAccountRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	err := ah.AuthService.ConfirmAccount(ctx, req.Email, req.Code)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	utils.Response(ctx, nil)
}
