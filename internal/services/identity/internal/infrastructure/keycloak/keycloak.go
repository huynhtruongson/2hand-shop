package keycloak

import (
	"context"

	"github.com/Nerzal/gocloak/v14"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
	svErr "github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/errors"
)

// KeycloakClient implements the auth.authProvider interface using gocloak.
type KeycloakClient struct {
	cfg    config.KeycloakConfig
	client *gocloak.GoCloak
}

// NewKeycloakClient creates a new KeycloakClient instance.
func NewKeycloakClient(cfg config.KeycloakConfig) *KeycloakClient {
	return &KeycloakClient{
		cfg:    cfg,
		client: gocloak.NewClient(cfg.BaseURL),
	}
}

// UpdateUserInternalID updates the internal ID of a user in Keycloak.
func (kp *KeycloakClient) UpdateUserInternalID(ctx context.Context, userId, internalId string) error {
	// Login as client to get admin token for creating users
	token, err := kp.client.LoginClient(ctx, kp.cfg.ClientID, kp.cfg.ClientSecret, kp.cfg.Realm)
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("KeycloakClient.UpdateUserInternalID: LoginClient")
	}

	user, err := kp.client.GetUserByID(ctx, token.AccessToken, kp.cfg.Realm, userId)
	if err != nil {
		return svErr.ErrInternal.WithCause(err).WithInternal("KeycloakClient.UpdateUserInternalID: GetUserByID")
	}

	if user == nil {
		return svErr.ErrUserNotFound.WithInternal("KeycloakClient.UpdateUserInternalID: user not found")
	}

	if user.Attributes == nil {
		user.Attributes = make(map[string][]string)
	}
	user.Attributes["internalUserId"] = []string{internalId}

	return kp.client.UpdateUser(ctx, token.AccessToken, kp.cfg.Realm, *user)

}
