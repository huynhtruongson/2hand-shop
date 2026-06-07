package dto

import "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Gender    string `json:"gender" binding:"required,oneof=male female"`
}

type GetProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
}

func ToGetProfileResponse(user *entity.User) GetProfileResponse {
	return GetProfileResponse{
		ID:        user.ID(),
		Email:     user.Email(),
		FirstName: user.FirstName(),
		LastName:  user.LastName(),
		Gender:    user.Gender(),
	}
}
