package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
	svErr "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"
)

type userModel struct {
	ID             string       `db:"id"`
	AuthProvider   string       `db:"auth_provider"`
	AuthProviderID string       `db:"auth_provider_id"`
	Email          string       `db:"email"`
	FirstName      string       `db:"first_name"`
	LastName       string       `db:"last_name"`
	Gender         string       `db:"gender"`
	Role           string       `db:"role"`
	EmailVerified  bool         `db:"email_verified"`
	CreatedAt      time.Time    `db:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at"`
	DeletedAt      sql.NullTime `db:"deleted_at"`
}

func toModel(u *entity.User) *userModel {
	return &userModel{
		ID:             u.ID(),
		AuthProvider:   u.AuthProvider(),
		AuthProviderID: u.AuthProviderID(),
		Email:          u.Email(),
		FirstName:      u.FirstName(),
		LastName:       u.LastName(),
		Gender:         u.Gender(),
		Role:           u.Role().String(),
		EmailVerified:  u.EmailVerified(),
		CreatedAt:      u.CreatedAt(),
		UpdatedAt:      u.UpdatedAt(),
		DeletedAt:      utils.TimePtrToNullTime(u.DeletedAt()),
	}
}
func (u userModel) toAggregate() (*entity.User, error) {
	role, err := valueobject.NewUserRoleFromString(u.Role)
	if err != nil {
		return nil, err
	}
	return entity.UnmarshalUserFromDB(
		u.ID,
		u.AuthProvider,
		u.AuthProviderID,
		u.Email,
		u.FirstName,
		u.LastName,
		u.Gender,
		u.EmailVerified,
		role,
		u.CreatedAt,
		u.UpdatedAt,
		utils.NullTimeToPtr(u.DeletedAt),
	), nil
}

type UserRepo struct {
	// db postgressqlx.DB
}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (ur *UserRepo) GetUserByEmail(ctx context.Context, db postgressqlx.Querier, email string) (*entity.User, error) {
	const query = `
		SELECT id, auth_provider, auth_provider_id, email, first_name, last_name, gender, email_verified, role, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1
		LIMIT 1`

	var model userModel
	err := db.QueryRowxContext(ctx, query, email).StructScan(&model)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svErr.ErrUserNotFound
		}
		return nil, svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.GetUserByEmail")
	}

	aggUser, err := model.toAggregate()
	if err != nil {
		return nil, svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.GetUserByEmail")
	}
	return aggUser, nil
}

func (ur *UserRepo) CreateUser(ctx context.Context, db postgressqlx.Querier, user *entity.User) error {
	const query = `
		INSERT INTO users (id, auth_provider, auth_provider_id, email, first_name, last_name, gender, role, email_verified, created_at, updated_at, deleted_at)
		VALUES (:id, :auth_provider, :auth_provider_id, :email, :first_name, :last_name, :gender, :role, :email_verified, :created_at, :updated_at, :deleted_at)`

	model := toModel(user)
	_, err := db.NamedExecContext(ctx, query, model)
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.CreateUser")
	}

	return nil
}

func (ur *UserRepo) UpdateUserVerified(ctx context.Context, db postgressqlx.Querier, email string) error {
	const query = `
		UPDATE users
		SET email_verified = $1, updated_at = $2
		WHERE email = $3`

	now := time.Now().UTC()
	result, err := db.ExecContext(ctx, query, true, now, email)
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.UpdateUserVerified")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.UpdateUserVerified")
	}
	if rows == 0 {
		return svErr.ErrUserNotFound
	}

	return nil
}

func (ur *UserRepo) GetUserByID(ctx context.Context, db postgressqlx.Querier, userID string) (*entity.User, error) {
	const query = `
		SELECT id, auth_provider, auth_provider_id, email, first_name, last_name, gender, role, email_verified, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
		LIMIT 1`

	var model userModel
	err := db.QueryRowxContext(ctx, query, userID).StructScan(&model)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svErr.ErrUserNotFound
		}
		return nil, svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.GetUserByID")
	}

	aggUser, err := model.toAggregate()
	if err != nil {
		return nil, svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.GetUserByID")
	}
	return aggUser, nil
}

func (ur *UserRepo) UpdateUserProfile(ctx context.Context, db postgressqlx.Querier, user *entity.User) error {
	const query = `
		UPDATE users
		SET first_name = $1, last_name = $2, gender = $3, updated_at = $4
		WHERE id = $5`

	result, err := db.ExecContext(ctx, query, user.FirstName(), user.LastName(), user.Gender(), time.Now(), user.ID())
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.UpdateUserProfile")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("UserRepo.UpdateUserProfile")
	}
	if rows == 0 {
		return svErr.ErrUserNotFound
	}

	return nil
}
