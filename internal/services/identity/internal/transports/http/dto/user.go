package dto

import "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"

type UpdateProfileRequest struct {
	Name   string `json:"name" binding:"required"`
	Gender string `json:"gender" binding:"required,oneof=male female"`
}

type GetProfileResponse struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

func ToGetProfileResponse(user *entity.User) GetProfileResponse {
	return GetProfileResponse{
		ID:     user.ID(),
		Email:  user.Email(),
		Name:   user.Name(),
		Gender: user.Gender(),
	}
}
