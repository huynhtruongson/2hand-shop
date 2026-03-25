package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
)

type UserRepo interface {
	GetUserByID(ctx context.Context, querier postgressqlx.Querier, userID string) (*entity.User, error)
	UpdateUserProfile(ctx context.Context, querier postgressqlx.Querier, user *entity.User) error
}
