package storage

import (
	"context"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage/driver"
	"github.com/thatique/kuade/pkg/kerr"
	"github.com/thatique/kuade/pkg/metrics"
)

// UserStore is storage for user
type UserStore struct {
	store  driver.UserStore
	tracer *metrics.Tracer
}

// NewUserStore create UserStore
func NewUserStore(store driver.UserStore) *UserStore {
	return &UserStore{
		store: store,
		tracer: &metrics.Tracer{
			Package:        pkgName,
			Provider:       metrics.ProviderName(store),
			LatencyMeasure: latencyMeasure,
		},
	}
}

// PutUser set user data or insert if it didn't exist
func (s *UserStore) PutUser(ctx context.Context, user *model.User) (err error) {
	ctx = s.tracer.Start(ctx, "PutUser")
	defer func() { s.tracer.End(ctx, err) }()

	err = s.store.PutUser(ctx, user)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

// PutUserCredential set user's credentials or create it if it doesn't exists
func (s *UserStore) PutUserCredential(ctx context.Context, id model.ID, creds *model.Credentials) (err error) {
	ctx = s.tracer.Start(ctx, "PutUserCredential")
	defer func() { s.tracer.End(ctx, err) }()

	err = s.store.PutUserCredential(ctx, id, creds)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

// FindOrCreateUserForProvider Find or create user based passed oauth provider
func (s *UserStore) FindOrCreateUserForProvider(ctx context.Context, userdata *model.User, provider model.OauthProvider) (newuser bool, user *model.User, err error) {
	ctx = s.tracer.Start(ctx, "FindOrCreateUserForProvider")
	defer func() { s.tracer.End(ctx, err) }()

	newuser, user, err = s.store.FindOrCreateUserForProvider(ctx, userdata, provider)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

// IsEmailAlreadyInUse check the passed email already in use or not
func (s *UserStore) IsEmailAlreadyInUse(ctx context.Context, email string) (used bool, id model.ID, err error) {
	ctx = s.tracer.Start(ctx, "IsEmailAlreadyInUse")
	defer func() { s.tracer.End(ctx, err) }()

	used, id, err = s.store.IsEmailAlreadyInUse(ctx, email)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

// GetCredentialByEmail get user credential by email
func (s *UserStore) GetCredentialByEmail(ctx context.Context, email string) (creds *model.Credentials, err error) {
	ctx = s.tracer.Start(ctx, "GetCredentialByEmail")
	defer func() { s.tracer.End(ctx, err) }()

	creds, err = s.store.GetCredentialByEmail(ctx, email)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

// GetUserByID get user by their ID
func (s *UserStore) GetUserByID(ctx context.Context, id model.ID) (user *model.User, err error) {
	ctx = s.tracer.Start(ctx, "GetUserByID")
	defer func() { s.tracer.End(ctx, err) }()

	user, err = s.store.GetUserByID(ctx, id)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

// GetUserBySlug get user by their slug
func (s *UserStore) GetUserBySlug(ctx context.Context, slug string) (user *model.User, err error) {
	ctx = s.tracer.Start(ctx, "GetUserBySlug")
	defer func() { s.tracer.End(ctx, err) }()

	user, err = s.store.GetUserBySlug(ctx, slug)
	if err != nil {
		err = wrapUserStoreError(s.store, err)
	}
	return
}

func wrapUserStoreError(s driver.UserStore, err error) error {
	if kerr.DoNotWrap(err) {
		return err
	}
	return kerr.New(s.ErrorCode(err), err, 2, "UserStore")
}
