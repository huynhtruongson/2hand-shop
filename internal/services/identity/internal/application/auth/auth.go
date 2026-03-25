package auth

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"

	appErr "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
)

type AuthService struct {
	logger       logger.Logger
	db           postgressqlx.DB
	authProvider authProvider
	userRepo     userRepo
}

type authProvider interface {
	ProviderName() string
	SignUp(ctx context.Context, email, password string, attrs map[string]string) (*AuthProviderSignUpResp, error)
	SignIn(ctx context.Context, email, password string) (*AuthProviderSignInResp, error)
	ConfirmAccount(ctx context.Context, email, code string) error
}

type userRepo interface {
	GetUserByEmail(ctx context.Context, db postgressqlx.Querier, email string) (*entity.User, error)
	CreateUser(ctx context.Context, db postgressqlx.Querier, user *entity.User) error
	UpdateUserVerified(ctx context.Context, db postgressqlx.Querier, email string) error
}

func NewAuthService(logger logger.Logger, db postgressqlx.DB, authProvider authProvider, userRepo userRepo) *AuthService {
	return &AuthService{
		logger:       logger,
		db:           db,
		authProvider: authProvider,
		userRepo:     userRepo,
	}
}

func (as *AuthService) SignUp(ctx context.Context, params SignUpParams) (*SignUpResult, error) {
	var result SignUpResult
	if err := postgressqlx.ExecTx(ctx, as.db, func(ctx context.Context, tx postgressqlx.TX) error {
		user, err := as.userRepo.GetUserByEmail(ctx, tx, params.Email)
		if err != nil && !appErr.IsKind(err, appErr.KindNotFound) {
			return err
		}
		if user != nil {
			return errors.ErrUserAlreadyExists
		}
		userID := utils.GenerateXID()

		out, err := as.authProvider.SignUp(ctx, params.Email, params.Password,
			map[string]string{
				"custom:id":   userID,
				"custom:role": valueobject.UserRoleClient.String(),
			})
		if err != nil {
			return err
		}

		user, err = entity.NewUser(userID, params.Email, params.Name, params.Gender, valueobject.UserRoleClient)
		if err != nil {
			return err
		}
		user.WithAuthProvider(as.authProvider.ProviderName(), out.UserSub)
		if err := as.userRepo.CreateUser(ctx, tx, user); err != nil {
			return err
		}
		result.UserID = userID
		result.IsVerified = out.IsVerified

		return nil
	}); err != nil {
		return nil, err
	}
	return &result, nil
}

func (as *AuthService) SignIn(ctx context.Context, email, password string) (*SignInResult, error) {
	user, err := as.userRepo.GetUserByEmail(ctx, as.db, email)
	if err != nil {
		if appErr.IsKind(err, appErr.KindNotFound) {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, err
	}
	if !user.IsVerified() {
		return nil, errors.ErrUserNotVerified
	}

	out, err := as.authProvider.SignIn(ctx, email, password)
	if err != nil {
		return nil, err
	}

	return &SignInResult{
		AccessToken:  out.AccessToken,
		IDToken:      out.IDToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    out.ExpiresIn,
	}, nil
}

func (as *AuthService) ConfirmAccount(ctx context.Context, email, code string) error {
	if err := postgressqlx.ExecTx(ctx, as.db, func(ctx context.Context, tx postgressqlx.TX) error {
		user, err := as.userRepo.GetUserByEmail(ctx, tx, email)
		if err != nil {
			return err
		}
		if err := as.authProvider.ConfirmAccount(ctx, email, code); err != nil {
			return err
		}
		if user.IsVerified() {
			return nil
		}
		return as.userRepo.UpdateUserVerified(ctx, tx, email)
	}); err != nil {
		return err
	}

	return nil
}
