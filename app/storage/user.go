package storage

import (
	"context"

	"github.com/thatique/kuade/api/v1"
	"github.com/thatique/kuade/app/model"
)

type UserStorage interface {
	FindUserById(ctx context.Context, id v1.ObjectID) (*model.User, error)

	FindUserByEmail(ctx context.Context, email string) (*model.User, error)

	// find a user by slug
	FindUserBySlug(ctx context.Context, slug string) (*model.User, error)

	FindOrCreateUserForProvider(ctx context.Context, userdata *model.User,
		provider model.OAuthProvider) (newuser bool, user *model.User, err error)

	InsertUser(ctx context.Context, user *model.User) (v1.ObjectID, error)

	UpdateUser(ctx context.Context, user *model.User) error

	// Update user credentials
	UpdateUserCredentials(context.Context, v1.ObjectID, model.Credentials) error

	// Update user's profile
	UpdateUserProfile(context.Context, v1.ObjectID, model.Profile) error

	// Delete
	DeleteUser(context.Context, *model.User) error
}
