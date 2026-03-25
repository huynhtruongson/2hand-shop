package query

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/repository"
)

type ProfileHandler cqrs.QueryHandler[Profile, *entity.User]

type Profile struct {
	ID string
}

type profileHandler struct {
	db       postgressqlx.DB
	userRepo repository.UserRepo
}

func NewProfileHandler(db postgressqlx.DB, userRepo repository.UserRepo) ProfileHandler {
	return &profileHandler{
		db:       db,
		userRepo: userRepo,
	}
}

func (h *profileHandler) Handle(ctx context.Context, query Profile) (*entity.User, error) {
	return h.userRepo.GetUserByID(ctx, h.db, query.ID)
}
