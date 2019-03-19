package passwords

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/auth/passwords/notifications"
	"github.com/thatique/kuade/kuade/auth/passwords/tokens"
	"github.com/thatique/kuade/pkg/emailparser"
)

var ErrInvalidEmail = errors.New("passwords: invalid email")

type ErrorCode int

const (
	NoError ErrorCode = iota
	ErrPassNotMatch
	ErrMinimumStaffPassLen
	ErrMinimumPassword
	ErrUpstream
)

func (code ErrorCode) ErrorDescription() string {
	switch code {
	case NoError:
		return "no error"
	case ErrPassNotMatch:
		return "password dan konfirmasi password tidak sama"
	case ErrMinimumStaffPassLen:
		return "panjang password harus lebih dari 15 karaketer"
	case ErrMinimumPassword:
		return "panjang password harus lebih dari 8 karakter"
	default:
		return "unknown error"
	}
}

type Broker struct {
	ResetURL string
	Message  string
	store    auth.UserStore
	token    tokens.TokenGenerator
	notifier notifications.Notifier
}

func NewBroker(t tokens.TokenGenerator, n notifications.Notifier, store auth.UserStore) *Broker {
	return &Broker{store: store, token: t, notifier: n}
}

type ResetRequest struct {
	user                        *auth.User
	token, Password1, Password2 string
}

func (r *ResetRequest) GetUser() *auth.User {
	return r.user
}

func (b *Broker) SendResetLink(ctx context.Context, ip string, email string) error {
	if !emailparser.IsValidEmail(email) {
		return ErrInvalidEmail
	}

	user, err := b.store.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	token, err := b.token.Generate(user)
	if err != nil {
		return err
	}

	uid := base64.RawURLEncoding.EncodeToString([]byte(user.Id))
	link := fmt.Sprintf("%s/%s/%s", b.ResetURL, uid, token)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		t := template.Must(template.New("reset").Parse(b.Message))
		if err := t.Execute(w, map[string]interface{}{
			"link":      link,
			"Email":     user.Email,
			"CreatedAt": time.Now().UTC().Format(time.RFC1123),
			"IP":        ip,
		}); err != nil {
			panic(err)
		}
	}()

	return b.notifier.Notify(user, r)
}

func (b *Broker) Resets(req *ResetRequest, fn func(user *auth.User, pswd string) error) ErrorCode {
	if req.Password1 != req.Password2 {
		return ErrPassNotMatch
	}

	if req.user.Role == auth.ROLE_STAFF && len(req.Password1) < 15 {
		return ErrMinimumStaffPassLen
	}

	if len(req.Password1) < 8 {
		return ErrMinimumPassword
	}

	// password check success, call the callback, then delete the token
	err := fn(req.user, req.Password1)
	if err != nil {
		return ErrUpstream
	}
	if err = b.token.Delete(req.token); err != nil {
		return ErrUpstream
	}

	return NoError
}

func (b *Broker) ValidateReset(ctx context.Context, uid, token string) (req *ResetRequest, ok bool) {
	if uid == "" || token == "" {
		ok = false
		return
	}

	objectid, ok := b.validateUid(uid)
	if !ok {
		ok = false
		return
	}

	user, err := b.store.FindById(ctx, objectid)
	if err != nil {
		ok = false
		return
	}

	req = &ResetRequest{user: user}

	ok = b.token.IsValid(user, token)
	if !ok {
		ok = false
		return
	}

	req.token = token
	ok = true
	return
}

func (b *Broker) validateUid(uid string) (bson.ObjectId, bool) {
	bs, err := base64.RawURLEncoding.DecodeString(uid)
	if err != nil {
		return bson.ObjectId(""), false
	}

	objectid := bson.ObjectId(bs[:])
	return objectid, objectid.Valid()
}
