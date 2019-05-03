package driver

import (
	"github.com/syaiful6/sersan"
)

// Driver is storage driver
type Driver interface {
	// GetSessionStore return sersan.Storage implementation
	GetSessionStore() (sersan.Storage, error)
	// GetUserStorage
	GetUserStore() (UserStore, error)
}
