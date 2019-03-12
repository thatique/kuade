package storage

import (
	"github.com/thatique/kuade/kuade/auth"
)

// Storage Driver
type Driver interface {
	Name() string
	// Get user storage
	GetUserStorage() (store auth.UserStore, err error)
}
