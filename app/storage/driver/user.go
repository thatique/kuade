package driver

import (
	"context"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/pkg/kerr"
)

// UserStore is storage for our identities management
type UserStore interface {
	// PutUser Insert a user if it not exists or update it
	PutUser(context.Context, *model.User) error

	// PutUserCredential set or create user credential
	PutUserCredential(context.Context, model.ID, *model.Credentials) error

	// FindOrCreateUserForProvider find an user for provided oauth provider
	// and return weather it's new user or not and user it'self
	FindOrCreateUserForProvider(context.Context, *model.User, model.OauthProvider) (bool, *model.User, error)

	// IsEmailAlreadyInUse check if the given username already in use, if so return
	// the ID of the user, if not return
	IsEmailAlreadyInUse(context.Context, string) (bool, model.ID, error)

	// IsUsernameAlreadyInUse check if the given username already in use, if so return
	// the ID of the user, if not return false
	IsUsernameAlreadyInUse(context.Context, string) (bool, model.ID, error)

	// GetCredentialByEmail get user credential by email
	GetCredentialByEmail(context.Context, string) (*model.Credentials, error)

	// GetCredentialByUsername get user credential by username
	GetCredentialByUsername(context.Context, string) (*model.Credentials, error)

	// GetUserByID get user by their ID
	GetUserByID(context.Context, model.ID) (*model.User, error)

	// GetUserBySlug get user by their slug
	GetUserBySlug(context.Context, string) (*model.User, error)

	// ErrorCode should return a code that describes the error, which was returned by
	// one of the other methods in this interface.
	ErrorCode(err error) kerr.ErrorCode
}
