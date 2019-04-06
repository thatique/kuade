package storage

import (
	"errors"
)

// Storage Error
var (
	// return this error if the requested object does not exist
	ErrNotFound = errors.New("Object doesn't exists")
)

type Driver interface {
	GetUserStorage() (UserStorage, error)
}
