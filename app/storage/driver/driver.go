package driver

import (
	"github.com/thatique/kuade/app/config"
)

// Driver is storage driver
type Driver interface {
	config.Configurable

	// GetUserStorage
	GetUserStorage() (UserStore, error)
}
