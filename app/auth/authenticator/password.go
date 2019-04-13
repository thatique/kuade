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
	errExpectedEmail     = errors.New("authenticator: expected an email")
	errInvalidCredential = errors.New("authenticator: invalid credential")
)

var _ authenticator.Password = &passwordAuthenticator{}

type passwordAuthenticator struct {
	users storage.UserStorage
}

func runHasher(pswd string) {
	var usr *model.User
	usr.SetPassword([]byte(pswd))
}

func (pswd *passwordAuthenticator) AuthenticatePassword(ctx context.Context, username, password string) (*authenticator.Response, bool, error) {
	if username == "" || !emailparser.IsValidEmail(username) {
		return nil, false, errExpectedEmail
	}

	usr, err := pswd.users.FindUserByEmail(ctx, username)
	if err != nil {
		runHasher(password)
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

	if !usr.VerifyPassword([]byte(password)) {
		return nil, false, errInvalidCredential
	}

	credentials := usr.Credentials
	credentials.LastSignin = time.Now().UTC()

	// update last login
	err = pswd.users.UpdateUserCredentials(context.Background(), usr.ID, credentials)
	if err != nil {
		return nil, false, err
	}

	return &authenticator.Response{User: usr.ToAuthInfo()}, true, nil
}
