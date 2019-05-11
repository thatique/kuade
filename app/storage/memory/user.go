package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/pkg/kerr"
)

type userStore struct {
	sync.RWMutex

	users  map[model.ID]*model.User
	bySlug map[string]model.ID

	creds           map[string]*model.Credentials
	credsByUsername map[string]string

	byProviders map[string]model.ID
}

func newUserStore() *userStore {
	return &userStore{
		users:       make(map[model.ID]*model.User),
		bySlug:      make(map[string]model.ID),
		creds:       make(map[string]*model.Credentials),
		byProviders: make(map[string]model.ID),
	}
}

func (u *userStore) PutUser(ctx context.Context, user *model.User) error {
	u.Lock()
	defer u.Unlock()

	u.users[user.ID] = user
	u.bySlug[user.Slug] = user.ID

	return nil
}

func (u *userStore) PutUserCredential(ctx context.Context, cred *model.Credentials) error {
	u.Lock()
	defer u.Unlock()

	u.creds[cred.GetEmail()] = cred
	u.credsByUsername[cred.GetUsername()] = cred.GetEmail()
	return nil
}

func (u *userStore) FindOrCreateUserForProvider(ctx context.Context, user *model.User, provider model.OauthProvider) (bool, *model.User, error) {
	u.Lock()

	if u.byProviders == nil {
		u.byProviders = make(map[string]model.ID)
	}

	key := fmt.Sprintf("%s:%s", provider.GetName(), provider.GetKey())
	if uid, ok := u.byProviders[key]; ok {
		if usr, ok := u.users[uid]; ok {
			u.Unlock()
			return false, usr, nil
		}
		u.Unlock()
		return false, nil, fmt.Errorf("errors provider name and key pointed to non existed user %d", int64(uid))
	}

	u.byProviders[key] = user.ID
	u.Unlock()
	u.PutUser(ctx, user)

	return true, user, nil
}

func (u *userStore) GetCredentialByEmail(ctx context.Context, email string) (*model.Credentials, error) {
	u.RLock()
	defer u.RUnlock()

	if cred, ok := u.creds[email]; ok {
		return cred, nil
	}

	return nil, errNotFound
}

func (u *userStore) GetCredentialByUsername(ctx context.Context, username string) (*model.Credentials, error) {
	u.RLock()
	defer u.RUnlock()

	if email, ok := u.credsByUsername[username]; ok {
		if cred, ok := u.creds[email]; ok {
			return cred, nil
		}
	}

	return nil, errNotFound
}

func (u *userStore) GetUserByID(ctx context.Context, id model.ID) (*model.User, error) {
	u.RLock()
	defer u.RUnlock()

	if user, ok := u.users[id]; ok {
		return user, nil
	}
	return nil, errNotFound
}

func (u *userStore) GetUserBySlug(ctx context.Context, slug string) (*model.User, error) {
	u.RLock()
	defer u.RUnlock()

	if id, ok := u.bySlug[slug]; ok {
		if user, ok := u.users[id]; ok {
			return user, nil
		}
	}

	return nil, errNotFound
}

func (u *userStore) IsEmailAlreadyInUse(ctx context.Context, email string) (bool, model.ID, error) {
	creds, err := u.GetCredentialByEmail(ctx, email)
	if err != nil {
		if err == errNotFound {
			err = nil
		}
		return false, model.ID(0), err
	}

	return true, creds.UserID, nil
}

func (u *userStore) IsUsernameAlreadyInUse(ctx context.Context, username string) (bool, model.ID, error) {
	creds, err := u.GetCredentialByUsername(ctx, username)
	if err != nil {
		if err == errNotFound {
			err = nil
		}
		return false, model.ID(0), err
	}

	return true, creds.UserID, nil
}

func (u *userStore) ErrorCode(err error) kerr.ErrorCode {
	if err == errNotFound {
		return kerr.NotFound
	}

	return kerr.Unknown
}
