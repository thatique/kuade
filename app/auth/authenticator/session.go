package authenticator

import (
	"errors"
	"net/http"

	"github.com/syaiful6/sersan"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
)

const (
	// UserSessionKey is key used to store user id in session
	UserSessionKey = "_userid_"
)

var (
	// ErrUserSessionDoesnotExist whenever we can't find an user in session
	ErrUserSessionDoesnotExist = errors.New("auth: no user in session")
	// ErrInvalidUserSessionID is thrown when we detect invalid user id
	ErrInvalidUserSessionID = errors.New("auth: invalid user session id")
)

// Session is authenticator
type Session struct {
	users *storage.UserStore
}

// NewSessionAuthenticator create user
func NewSessionAuthenticator(store *storage.UserStore) *Session {
	return &Session{users: store}
}

// Login login user to application and store it in session until it expired
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

// Logout logout user from application, remove userInfoContext if it can
func Logout(r *http.Request) error {
	session, err := sersan.GetSession(r)
	if err != nil {
		return err
	}

	delete(session, UserSessionKey)

	return nil
}

// AuthenticateRequest authenticate user by request information
func (sess *Session) AuthenticateRequest(r *http.Request) (*authenticator.Response, bool, error) {
	user, err := sess.loadUserFromSession(r)
	if err != nil {
		if err == ErrUserSessionDoesnotExist {
			session, err := sersan.GetSession(r)
			if err == nil {
				delete(session, UserSessionKey)
			}
		}
		return nil, false, err
	}

	response := &authenticator.Response{
		User: user,
	}
	return response, true, nil
}

// update session
func updateSession(r *http.Request, u *model.User) error {
	sessionMap, err := sersan.GetSession(r)
	if err != nil {
		return err
	}

	sessionMap[UserSessionKey] = u.ID.String()
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
		return nil, ErrUserSessionDoesnotExist
	}

	if uid, ok = suid.(string); !ok {
		return nil, ErrInvalidUserSessionID
	}

	objectid, err := model.NewIDFromString(uid)
	if err != nil {
		return nil, err
	}

	user, err := sess.users.GetUserByID(r.Context(), objectid)
	if err != nil {
		return nil, err
	}

	return user, nil
}
