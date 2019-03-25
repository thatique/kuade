package storage

import (
	"errors"
	"github.com/thatique/kuade/kuade/auth"
)

// Storage Error
var (
	// return this error if the requested object does not exist
	ErrNotFound = errors.New("Object doesn't exists")
)

// Storage Driver
type Driver interface {
	Name() string
	// Get user storage
	GetUserStorage() (store auth.UserStore, err error)
}
