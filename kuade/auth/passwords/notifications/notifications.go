package notifications

import (
	"io"

	"github.com/thatique/kuade/kuade/auth"
)

type Notifier interface {
	// notify message to the give user
	Notify(user *auth.User, message io.Reader) error
}
