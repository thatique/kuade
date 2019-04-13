package authenticator

import (
	"context"
	"net/http"

	"github.com/thatique/kuade/pkg/iam/auth/user"
)

// Token checks a string value against a backing authentication store and
// returns a Response or an error if the token could not be checked.
type Token interface {
	AuthenticateToken(ctx context.Context, token string) (*Response, bool, error)
}

// Password authenticator can authenticate user using username and password pair
// the username parameter can be an email or anything.
type Password interface {
	AuthenticatePassword(ctx context.Context, username, password string) (*Response, bool, error)
}

// Request attempts to extract authentication information from a request and
// returns a Response or an error if the request could not be checked.
type Request interface {
	AuthenticateRequest(req *http.Request) (*Response, bool, error)
}

// TokenFunc is a function that implements the Token interface.
type TokenFunc func(ctx context.Context, token string) (*Response, bool, error)

// AuthenticateToken implements authenticator.Token.
func (f TokenFunc) AuthenticateToken(ctx context.Context, token string) (*Response, bool, error) {
	return f(ctx, token)
}

// RequestFunc is a function that implements the Request interface.
type RequestFunc func(req *http.Request) (*Response, bool, error)

// AuthenticateRequest implements authenticator.Request.
func (f RequestFunc) AuthenticateRequest(req *http.Request) (*Response, bool, error) {
	return f(req)
}

// PasswordFunc is a function that implements the Password interface.
type PasswordFunc func(ctx context.Context, user, password string) (*Response, bool, error)

// AuthenticatePassword implements authenticator.Password.
func (f PasswordFunc) AuthenticatePassword(ctx context.Context, user, password string) (*Response, bool, error) {
	return f(ctx, user, password)
}

type Response struct {
	// Audiences is the set of audiences the authenticator was able to validate
	// the token against. If the authenticator is not audience aware, this field
	// will be empty.
	Audiences Audiences
	// User is the UserInfo associated with the authentication context.
	User user.Info
}
