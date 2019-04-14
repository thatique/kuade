package storage

import (
	"context"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage/driver"
	"github.com/thatique/kuade/pkg/kerr"
	"github.com/thatique/kuade/pkg/metrics"
)

// UserStorage is storage for user
type UserStorage struct {
	storage driver.UserStorage
	tracer  *metrics.Tracer
}

// NewUserStorage create UserStorage
func NewUserStorage(storage driver.UserStorage) *UserStorage {
	return &UserStorage{
		storage: storage,
		tracer: &metrics.Tracer{
			Package:        pkgName,
			Provider:       metrics.ProviderName(storage),
			LatencyMeasure: latencyMeasure,
		},
	}
}

// PutUser set user data or insert if it didn't exist
func (s *UserStorage) PutUser(ctx context.Context, user *model.User) (err error) {
	ctx = s.tracer.Start(ctx, "PutUser")
	defer func() { s.tracer.End(ctx, err) }()

	err = s.storage.PutUser(ctx, user)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// PutUserProfile set user's profile or create it if it doesn't exists
func (s *UserStorage) PutUserProfile(ctx context.Context, id model.ID, profile *model.Profile) (err error) {
	ctx = s.tracer.Start(ctx, "PutUserProfile")
	defer func() { s.tracer.End(ctx, err) }()

	err = s.storage.PutUserProfile(ctx, id, profile)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// PutUserCredential set user's credentials or create it if it doesn't exists
func (s *UserStorage) PutUserCredential(ctx context.Context, creds *model.Credentials) (err error) {
	ctx = s.tracer.Start(ctx, "PutUserCredential")
	defer func() { s.tracer.End(ctx, err) }()

	err = s.storage.PutUserCredential(ctx, id, creds)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// FindOrCreateUserForProvider Find or create user based passed oauth provider
func (s *UserStorage) FindOrCreateUserForProvider(ctx context.Context, userdata *model.User, provider model.OAuthProvider) (newuser bool, user *model.User, err error) {
	ctx = s.tracer.Start(ctx, "FindOrCreateUserForProvider")
	defer func() { s.tracer.End(ctx, err) }()

	newuser, user, err = s.storage.FindOrCreateUserForProvider(ctx, userdata, provider)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// GetCredentialByEmail get user credential by email
func (s *Storage) GetCredentialByEmail(ctx context.Context, email string) (creds *model.PasswordCredential, err error) {
	ctx = s.tracer.Start(ctx, "GetCredentialByEmail")
	defer func() { s.tracer.End(ctx, err) }()

	creds, err = s.stroge.GetCredentialByEmail(ctx, email)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// GetUserProfile get user's profile by user ID
func (s *Storage) GetUserProfile(ctx context.Context, id model.ID) (profile *model.Profile, err error) {
	ctx = s.tracer.Start(ctx, "GetUserProfile")
	defer func() { s.tracer.End(ctx, err) }()

	profile, err = s.storage.GetUserProfile(ctx, id)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// GetUserByID get user by their ID
func (s *Storage) GetUserByID(ctx context.Context, id model.ID) (user *model.User, err error) {
	ctx = s.tracer.Start(ctx, "GetUserByID")
	defer func() { s.tracer.End(ctx, err) }()

	user, err = s.storage.GetUserByID(ctx, id)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

// GetUserBySlug get user by their slug
func (s *Storage) GetUserBySlug(ctx context.Context, slug string) (user *model.User, err error) {
	ctx = s.tracer.Start(ctx, "GetUserBySlug")
	defer func() { s.tracer.End(ctx, err) }()

	user, err = s.storage.GetUserBySlug(ctx, slug)
	if err != nil {
		err = wrapUserStorageError(s.storage, err)
	}
	return
}

func wrapUserStorageError(s driver.UserStorage, err error) error {
	if kerr.DoNotWrap(err) {
		return err
	}
	return kerr.New(s.ErrorCode(err), err, 2, "userstorage")
}
