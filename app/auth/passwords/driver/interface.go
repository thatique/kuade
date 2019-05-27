package driver

import (
	"context"

	"github.com/thatique/kuade/app/model"
)

// ResetTokenGenerator used to generate and check tokens for the password
// reset mechanism
type ResetTokenGenerator interface {
	// Create return a token that can be used once to do a password reset
	// for the given user's credentials.
	Create(ctx context.Context, cred *model.Credentials) (string, error)

	// Check that a password reset token is correct for a given user.
	Check(ctx context.Context, user *model.Credentials, token string) bool
}
