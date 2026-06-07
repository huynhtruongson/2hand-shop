package application

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/eventhandler"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/query"
)

type Application struct {
	Commands      Commands
	Queries       Queries
	EventHandlers EventHandlers
}

type Commands struct {
	UpdateProfile command.UpdateProfileHandler
}

type Queries struct {
	Profile query.ProfileHandler
}

type EventHandlers struct {
	OnKeycloakUserRegistration eventhandler.OnKeycloakUserRegistrationHandler
}
