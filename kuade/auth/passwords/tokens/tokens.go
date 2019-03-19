package tokens

import (
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/pkg/text"
)

// TokenGenerator can create and check a token to be used in passwords
// reset workflow
type TokenGenerator interface {
	// Generate return a token that can be used once to do a password
	// reset for the given user
	Generate(user *auth.User) (token string, err error)

	// Delete this token, or make this token as invalid to be used again.
	// After this methond called, calling the same token to `IsValid) method
	// will result false.
	Delete(token string) error

	// IsValid check that a password reset token is valid
	IsValid(user *auth.User, token string) bool
}

type PasswordToken struct {
	Token     string
	Email     string
	Pass      []byte
	CreatedAt int64
}

const TokenAllowedChars = text.ASCII_LOWERCASE + text.ASCII_UPPERCASE + text.DIGITS + "-_~"

func GenerateToken() (string, error) {
	return text.RandomString(32, TokenAllowedChars)
}
