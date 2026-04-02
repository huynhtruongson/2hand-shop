package command

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/repository"
)

type UpdateProfileHandler cqrs.CommandHandler[UpdateProfile, UpdateProfileResponse]

type UpdateProfile struct {
	UserID string
	Name   string
	Gender string
}
type UpdateProfileResponse struct{}
type updateProfileHandler struct {
	db       postgressqlx.DB
	userRepo repository.UserRepo
}

func NewUpdateProfileHandler(db postgressqlx.DB, userRepo repository.UserRepo) UpdateProfileHandler {
	return &updateProfileHandler{
		db:       db,
		userRepo: userRepo,
	}
}
func (h *updateProfileHandler) Handle(ctx context.Context, cmd UpdateProfile) (UpdateProfileResponse, error) {
	user, err := h.userRepo.GetUserByID(ctx, h.db, cmd.UserID)
	if err != nil {
		return UpdateProfileResponse{}, err
	}
	if err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		user.UpdateProfile(cmd.Name, cmd.Gender)
		return h.userRepo.UpdateUserProfile(ctx, tx, user)
	}); err != nil {
		return UpdateProfileResponse{}, err
	}
	return UpdateProfileResponse{}, nil
}
