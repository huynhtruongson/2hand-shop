package command

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/repository"
)

type UpdateProfileHandler cqrs.CommandHandler[UpdateProfile, cqrs.CommandVoidResponse]

type UpdateProfile struct {
	UserID string
	Name   string
	Gender string
}
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
func (h *updateProfileHandler) Handle(ctx context.Context, cmd UpdateProfile) (cqrs.CommandVoidResponse, error) {
	user, err := h.userRepo.GetUserByID(ctx, h.db, cmd.UserID)
	if err != nil {
		return cqrs.CommandVoidResponse{}, err
	}
	if err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		user.UpdateProfile(cmd.Name, cmd.Gender)
		return h.userRepo.UpdateUserProfile(ctx, tx, user)
	}); err != nil {
		return cqrs.CommandVoidResponse{}, err
	}
	return cqrs.CommandVoidResponse{}, nil
}
