package authenticator

import (
	"errors"
	"net/http"

	"github.com/syaiful6/sersan"
	"github.com/thatique/kuade/api/v1"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/pkg/auth/authenticator"
)

const (
	UserSessionKey = "_userid_"
)

var (
	UserSessionDoesnotExist = errors.New("auth: no user in session")
	InvalidUserSessionId    = errors.New("auth: invalid user session id")
)

type Session struct {
	users storage.UserStorage
}

func NewSessionAuthenticator(store storage.UserStorage) *Session {
	return &Session{users: store}
}

// login user to application and store it in session until it expired
func Login(r *http.Request, u *model.User) error {
	if u == nil {
		return errors.New("user passed to login can't be nil user")
	}

	err := updateSession(r, u)
	if err != nil {
		return err
	}

	return nil
}

// logout user from application, remove userInfoContext if it can
func Logout(r *http.Request) error {
	session, err := sersan.GetSession(r)
	if err != nil {
		return err
	}

	delete(session, UserSessionKey)

	return nil
}

func (sess *Session) AuthenticateRequest(r *http.Request) (*authenticator.Response, bool, error) {
	user, err := sess.loadUserFromSession(r)
	if err != nil {
		if err == UserSessionDoesnotExist {
			session, err := sersan.GetSession(r)
			if err == nil {
				delete(session, UserSessionKey)
			}
		}
		return nil, false, err
	}

	response := &authenticator.Response{
		User: user.ToAuthInfo(),
	}
	return response, true, nil
}

// update session
func updateSession(r *http.Request, u *model.User) error {
	sessionMap, err := sersan.GetSession(r)
	if err != nil {
		return err
	}

	sessionMap[UserSessionKey] = u.ID.Hex()
	return nil
}

func (sess *Session) loadUserFromSession(r *http.Request) (*model.User, error) {
	sessionMap, err := sersan.GetSession(r)
	if err != nil {
		return nil, err
	}

	var (
		ok   bool
		uid  string
		suid interface{}
	)

	if suid, ok = sessionMap[UserSessionKey]; !ok {
		return nil, UserSessionDoesnotExist
	}

	if uid, ok = suid.(string); !ok {
		return nil, InvalidUserSessionId
	}

	objectid, err := v1.ObjectIDFromHex(uid)
	if err != nil {
		return nil, err
	}

	user, err := sess.users.FindUserById(r.Context(), objectid)
	if err != nil {
		return nil, err
	}

	return user, nil
}
