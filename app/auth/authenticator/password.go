package authenticator

import (
	"context"
	"errors"
	"time"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/pkg/emailparser"
	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
)

var (
	errExpectedEmailOrUsername = errors.New("authenticator: expected an email or username")
	errInvalidCredential       = errors.New("authenticator: invalid credential")
)

var _ authenticator.Password = &passwordAuthenticator{}

type passwordAuthenticator struct {
	users *storage.UserStore
}

// NewPasswordAuthenticator create a `authenticator.Password`
func NewPasswordAuthenticator(users *storage.UserStore) authenticator.Password {
	return &passwordAuthenticator{users: users}
}

func (pswd *passwordAuthenticator) AuthenticatePassword(ctx context.Context, username, password string) (*authenticator.Response, bool, error) {
	if username == "" {
		return nil, false, errExpectedEmailOrUsername
	}

	var credsFunc func(context.Context, string) (*model.Credentials, error)
	if emailparser.IsValidEmail(username) {
		credsFunc = pswd.users.GetCredentialByEmail
	} else {
		credsFunc = pswd.users.GetCredentialByUsername
	}

	creds, err := credsFunc(ctx, username)
	if err != nil {
		runHasher(password)
		return nil, false, err
	}

	if !creds.VerifyPassword([]byte(password)) {
		return nil, false, errInvalidCredential
	}

	// user credential passed, get the user
	usr, err := pswd.users.GetUserByID(ctx, creds.UserID)
	if err != nil {
		return nil, false, err
	}

	if !usr.IsActive() {
		var msg string
		if usr.Status == model.UserStatus_INACTIVE {
			msg = "status akun anda tidak aktif"
		} else {
			msg = "status akun anda sedang terkunci"
		}
		runHasher(password)
		return nil, false, errors.New(msg)
	}

	creds.LastSignin = time.Now().UTC()

	// update last login
	err = pswd.users.PutUserCredential(context.Background(), creds)
	if err != nil {
		return nil, false, err
	}

	return &authenticator.Response{User: usr}, true, nil
}

func runHasher(pswd string) {
	var usr *model.Credentials
	usr.SetPassword([]byte(pswd))
}
