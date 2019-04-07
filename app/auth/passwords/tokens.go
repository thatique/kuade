package passwords

import (
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/pkg/text"
)

type TokenGenerator interface {
	// Generate return a token that can be used once to do a password reset
	// for the given user.
	Generate(user *model.User) (token string, err error)

	// Delete this token, or make this token as invalid to be used again. After
	// this method called, calling the same token to `IsValid` method will result
	// false
	Delete(token string) error

	// IsValid check that a password reset token is valid
	IsValid(user *model.User, token string) bool
}

const TokenAllowedChars = text.ASCII_LOWERCASE + text.ASCII_UPPERCASE + text.DIGITS + "-_~"

func GenerateToken() (string, error) {
	return text.RandomString(32, TokenAllowedChars)
}
