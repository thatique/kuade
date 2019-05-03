package handlers

import (
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
)

type App struct {
	storage *storage.Store
}
