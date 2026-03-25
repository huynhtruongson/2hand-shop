package dto

import "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/auth"

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Gender   string `json:"gender" binding:"required,oneof=male female"`
}
type SignUpResp struct {
	UserID     string `json:"user_id"`
	IsVerified bool   `json:"is_verified"`
}

func (req SignUpRequest) ToSignUpParams() auth.SignUpParams {
	return auth.SignUpParams{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
		Gender:   req.Gender,
	}
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type SignInResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

type ConfirmAccountRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}
