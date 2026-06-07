package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
)

type AuthService struct {
	logger      logger.Logger
	keycloakCfg config.KeycloakConfig
	keycCache   *keyCache
}

type KeycloakClaims struct {
	jwt.RegisteredClaims
	PreferredUsername string              `json:"preferred_username"`
	Email             string              `json:"email"`
	EmailVerified     bool                `json:"email_verified"`
	Name              string              `json:"name"`
	InternalUserID    string              `json:"internal_user_id"`
	RealmAccess       realmAccess         `json:"realm_access"`
	ResourceAccess    map[string]roleList `json:"resource_access"`
	Scope             string              `json:"scope"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

// roleList holds per-resource roles embedded in resource_access.
type roleList struct {
	Roles []string `json:"roles"`
}

// type authProvider interface {
// 	ProviderName() string
// 	SignUp(ctx context.Context, param SignUpParams) (*AuthProviderSignUpResp, error)
// 	SignIn(ctx context.Context, email, password string) (*AuthProviderSignInResp, error)
// 	ConfirmAccount(ctx context.Context, email, code string) error
// }

// type userRepo interface {
// 	GetUserByEmail(ctx context.Context, db postgressqlx.Querier, email string) (*entity.User, error)
// 	CreateUser(ctx context.Context, db postgressqlx.Querier, user *entity.User) error
// 	UpdateUserVerified(ctx context.Context, db postgressqlx.Querier, email string) error
// }

func NewAuthService(logger logger.Logger, keycloakCfg config.KeycloakConfig) *AuthService {
	return &AuthService{
		logger:      logger,
		keycloakCfg: keycloakCfg,
		keycCache:   newKeyCache(),
	}
}

func (as *AuthService) VerifyToken(ctx *gin.Context) {
	tokenStr, err := extractBearer(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": err.Error()})
		return
	}
	claims, err := parseAndValidate(tokenStr, as.keycloakCfg, as.keycCache)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": err.Error()})
		return
	}

	ctx.Header(auth.HeaderUserID, claims.InternalUserID)
	ctx.Header(auth.HeaderEmail, claims.Email)
	ctx.Header(auth.HeaderRoles, strings.Join(claims.RealmAccess.Roles, ","))

	ctx.Status(http.StatusOK)

}

func parseAndValidate(tokenStr string, cfg config.KeycloakConfig, kc *keyCache) (*KeycloakClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&KeycloakClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing kid in token header")
			}
			return kc.get(kid, jwksURL(cfg))
		},
		// jwt.WithIssuer(issuerURL(cfg)),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func extractBearer(c *gin.Context) (string, error) {
	h := c.GetHeader("Authorization")
	if h == "" {
		return "", errors.New("no authorization header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("authorization header must be: Bearer <token>")
	}
	t := strings.TrimSpace(parts[1])
	if t == "" {
		return "", errors.New("empty bearer token")
	}
	return t, nil
}

func issuerURL(cfg config.KeycloakConfig) string {
	return fmt.Sprintf("%s/realms/%s", cfg.BaseURL, cfg.Realm)
}

func jwksURL(cfg config.KeycloakConfig) string {
	return issuerURL(cfg) + "/protocol/openid-connect/certs"
}

// func (as *AuthService) SignUp(ctx context.Context, params SignUpParams) (*SignUpResult, error) {
// 	var result SignUpResult

// 	if err := postgressqlx.ExecTx(ctx, as.db, func(ctx context.Context, tx postgressqlx.TX) error {
// 		user, err := as.userRepo.GetUserByEmail(ctx, tx, params.Email)
// 		if err != nil && !appErr.IsKind(err, appErr.KindNotFound) {
// 			return err
// 		}
// 		if user != nil {
// 			return errors.ErrUserAlreadyExists
// 		}
// 		userID := uuid.NewString()

// 		out, err := as.authProvider.SignUp(ctx, params)
// 		if err != nil {
// 			return err
// 		}

// 		user, err = entity.NewUser(userID, params.Email, params.FirstName, params.LastName, params.Gender, valueobject.UserRoleClient)
// 		if err != nil {
// 			return err
// 		}
// 		user.WithAuthProvider(as.authProvider.ProviderName(), out.UserSub)
// 		if err := as.userRepo.CreateUser(ctx, tx, user); err != nil {
// 			return err
// 		}
// 		result.UserID = userID
// 		result.IsVerified = out.IsVerified

// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}
// 	return &result, nil
// }

// func (as *AuthService) SignIn(ctx context.Context, email, password string) (*SignInResult, error) {
// 	user, err := as.userRepo.GetUserByEmail(ctx, as.db, email)
// 	if err != nil {
// 		if appErr.IsKind(err, appErr.KindNotFound) {
// 			return nil, errors.ErrInvalidCredentials
// 		}
// 		return nil, err
// 	}
// 	if !user.IsVerified() {
// 		return nil, errors.ErrUserNotVerified
// 	}

// 	out, err := as.authProvider.SignIn(ctx, email, password)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &SignInResult{
// 		AccessToken:  out.AccessToken,
// 		IDToken:      out.IDToken,
// 		RefreshToken: out.RefreshToken,
// 		ExpiresIn:    out.ExpiresIn,
// 	}, nil
// }

// func (as *AuthService) ConfirmAccount(ctx context.Context, email, code string) error {
// 	if err := postgressqlx.ExecTx(ctx, as.db, func(ctx context.Context, tx postgressqlx.TX) error {
// 		user, err := as.userRepo.GetUserByEmail(ctx, tx, email)
// 		if err != nil {
// 			return err
// 		}
// 		if err := as.authProvider.ConfirmAccount(ctx, email, code); err != nil {
// 			return err
// 		}
// 		if user.IsVerified() {
// 			return nil
// 		}
// 		return as.userRepo.UpdateUserVerified(ctx, tx, email)
// 	}); err != nil {
// 		return err
// 	}

// 	return nil
// }
