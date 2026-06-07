package eventhandler

import (
	"context"

	"github.com/google/uuid"
	appErr "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"
)

type OnKeycloakUserRegistrationHandler = dispatcher.TypedHandler[KeycloakUserRegistrationPayload]

type KeycloakClient interface {
	UpdateUserInternalID(ctx context.Context, userId, internalId string) error
}

type onKeycloakUserRegistrationHandler struct {
	logger   logger.Logger
	db       postgressqlx.DB
	keycloak KeycloakClient
	userRepo repository.UserRepo
}

func NewOnKeycloakUserRegistrationHandler(logger logger.Logger, db postgressqlx.DB, keycloak KeycloakClient, userRepo repository.UserRepo) OnKeycloakUserRegistrationHandler {
	return &onKeycloakUserRegistrationHandler{
		logger:   logger,
		db:       db,
		keycloak: keycloak,
		userRepo: userRepo,
	}
}

func (h *onKeycloakUserRegistrationHandler) Handle(ctx context.Context, ec types.EventContext[KeycloakUserRegistrationPayload]) error {
	userData := ec.Payload().Details

	user, err := h.userRepo.GetUserByEmail(ctx, h.db, userData.Email)
	if err != nil {
		if !appErr.IsKind(err, appErr.KindNotFound) {
			return err
		}
	}
	if user != nil {
		h.logger.Info("User already exists, skipping creation", "email", userData.Email)
		return nil
	}
	userID := uuid.NewString()
	user, err = entity.NewUser(userID, userData.Email, userData.FirstName, userData.LastName, "male", valueobject.UserRoleClient)
	if err != nil {
		return err
	}
	user.WithAuthProvider("keycloak", ec.Payload().UserID)

	if err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		if err := h.userRepo.CreateUser(ctx, tx, user); err != nil {
			return err
		}
		return h.keycloak.UpdateUserInternalID(ctx, ec.Payload().UserID, userID)
	}); err != nil {
		return err
	}
	return nil
}

type KeycloakUserRegistrationPayload struct {
	Class     string      `json:"@class"`
	Time      int64       `json:"time"`
	Type      string      `json:"type"`
	RealmID   string      `json:"realmId"`
	ClientID  string      `json:"clientId"`
	UserID    string      `json:"userId"`
	IPAddress string      `json:"ipAddress"`
	Details   UserDetails `json:"details"`
}

type UserDetails struct {
	AuthMethod     string `json:"auth_method"`
	AuthType       string `json:"auth_type"`
	RegisterMethod string `json:"register_method"`
	LastName       string `json:"last_name"`
	RedirectURI    string `json:"redirect_uri"`
	FirstName      string `json:"first_name"`
	CodeID         string `json:"code_id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
}
