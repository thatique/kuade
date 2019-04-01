package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/syaiful6/sersan"
	bson "go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	UserSessionKey = "_userid_"
	userCtxKey     = "auth.user"
	userCtxIdKey   = "auth.user.id"
)

var (
	UserSessionDoesnotExist = errors.New("auth: no user in session")
	InvalidUserSessionId    = errors.New("auth: invalid user session id")
)

type Session struct {
	users UserStore
}

func NewUserSession(store UserStore) *Session {
	return &Session{users: store}
}

func withContext(ctx context.Context, user *User) context.Context {
	return userInfoContext{
		Context: ctx,
		user:    user,
	}
}

type userInfoContext struct {
	context.Context
	user *User
}

func (uic userInfoContext) Value(key interface{}) interface{} {
	switch key {
	case userCtxKey:
		return uic.user
	case userCtxIdKey:
		return uic.user.Id
	}

	return uic.Context.Value(key)
}

// Middleware that load user from session and set it current user if success
func (sess *Session) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := sess.loadUserFromSession(r)
		if err != nil {
			if err == UserSessionDoesnotExist {
				session, err := sersan.GetSession(r)
				if err == nil {
					delete(session, UserSessionKey)
				}
			}
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, sess.LoginOnce(r, user))
	})
}

// get current user, return user if any
func (sess *Session) User(r *http.Request) *User {
	u, ok := r.Context().Value(userCtxKey).(*User)
	if !ok {
		return nil
	}
	return u
}

// login once set user only to current request without persisting to session store
func (sess *Session) LoginOnce(r *http.Request, u *User) *http.Request {
	return r.WithContext(withContext(r.Context(), u))
}

// login user to application and store it in session until it expired
func (sess *Session) Login(r *http.Request, u *User) (*http.Request, error) {
	if u == nil {
		return r, errors.New("user passed to login can't be nil user")
	}
	r = sess.LoginOnce(r, u)
	err := sess.updateSession(r, u)
	if err != nil {
		return r, err
	}

	return r, nil
}

// logout user from application, remove userInfoContext if it can
func (sess *Session) Logout(r *http.Request) (*http.Request, error) {
	session, err := sersan.GetSession(r)
	if err != nil {
		return nil, err
	}

	delete(session, UserSessionKey)

	// remove user from context
	ctx, ok := r.Context().(userInfoContext)
	if !ok {
		return r, nil
	}

	return r.WithContext(ctx.Context), nil
}

// update session
func (sess *Session) updateSession(r *http.Request, u *User) error {
	sessionMap, err := sersan.GetSession(r)
	if err != nil {
		return err
	}

	sessionMap[UserSessionKey] = u.Id.Hex()
	return nil
}

func (sess *Session) loadUserFromSession(r *http.Request) (*User, error) {
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

	objectid, err := bson.ObjectIDFromHex(uid)
	if err != nil {
		return nil, err
	}

	user, err := sess.users.FindById(r.Context(), objectid)
	if err != nil {
		return nil, err
	}

	return user, nil
}
