package application

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	UpdateProfile command.UpdateProfileHandler
}

type Queries struct {
	Profile query.ProfileHandler
}
