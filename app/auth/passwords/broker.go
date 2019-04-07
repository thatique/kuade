package passwords

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	htemplate "html/template"
	"net/mail"
	"text/template"
	"time"

	"github.com/emersion/go-message"
	"github.com/thatique/kuade/api/v1"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/pkg/emailparser"
	"github.com/thatique/kuade/pkg/mailer"
	"github.com/thatique/kuade/pkg/queue"
)

var ErrInvalidEmail = errors.New("passwords: invalid email")

type ErrorCode int32

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
		return "No Error"
	case ErrPassNotMatch:
		return "password and password confirmation not match"
	case ErrMinimumStaffPassLen:
		return "minimum password length for staff user is 15 chars"
	case ErrMinimumPassword:
		return "minimum password length is 8 chars"
	default:
		return "unknown error"
	}
}

type Broker struct {
	Sender, ResetURL, TemplatePath string
	Asset                          func(string) ([]byte, error)

	users     storage.UserStorage
	token     TokenGenerator
	transport *mailer.Transport
	queue     *queue.Queue
}

func NewBroker(users storage.UserStorage, token TokenGenerator, m *mailer.Transport, q *queue.Queue) *Broker {
	return &Broker{
		users:     users,
		token:     token,
		transport: m,
		queue:     q,
	}
}

type ResetRequest struct {
	user                        *model.User
	token, Password1, Password2 string
}

func (r *ResetRequest) GetUser() *model.User {
	return r.user
}

func (b *Broker) SendResetLink(ctx context.Context, ip string, email string) error {
	if !emailparser.IsValidEmail(email) {
		return ErrInvalidEmail
	}

	user, err := b.users.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	token, err := b.token.Generate(user)
	if err != nil {
		return err
	}

	uid := base64.RawURLEncoding.EncodeToString(user.ID[:])
	link := fmt.Sprintf("%s/%s/%s", b.ResetURL, uid, token)

	tplContext := map[string]interface{}{
		"link":      link,
		"Email":     user.Email,
		"CreatedAt": time.Now().UTC().Format(time.RFC1123),
		"IP":        ip,
	}

	// send this
	return b.notify(user, tplContext)
}

func (b *Broker) Resets(req *ResetRequest, fn func(user *model.User, pswd string) error) ErrorCode {
	if req.Password1 != req.Password2 {
		return ErrPassNotMatch
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

	objectid, ok := b.validateUserID(uid)
	if !ok {
		ok = false
		return
	}

	user, err := b.users.FindUserById(ctx, objectid)
	if err != nil {
		ok = false
		return
	}

	ok = b.token.IsValid(user, token)
	if !ok {
		ok = false
		return
	}

	req = &ResetRequest{user: user, token: token}
	ok = true
	return
}

func (b *Broker) validateUserID(uid string) (v1.ObjectID, bool) {
	bs, err := base64.RawURLEncoding.DecodeString(uid)
	if err != nil {
		return v1.NilObjectID, false
	}
	if len(bs) != 12 {
		return v1.NilObjectID, false
	}
	var oid [12]byte
	copy(oid[:], bs[:])
	return oid, true
}

func (b *Broker) notify(user *model.User, ctx map[string]interface{}) error {
	plainTpl, err := b.Asset(b.TemplatePath + ".txt")
	if err != nil {
		return err
	}
	t := template.Must(template.New("reset").Parse(string(plainTpl)))
	var buf bytes.Buffer
	if err = t.Execute(&buf, ctx); err != nil {
		return err
	}

	htmlTpl, err := b.Asset(b.TemplatePath + ".html")
	if err != nil {
		return err
	}
	ht := htemplate.Must(htemplate.New("reset").Parse(string(htmlTpl)))
	var buf2 bytes.Buffer
	if err = ht.Execute(&buf2, ctx); err != nil {
		return err
	}

	h1 := make(message.Header)
	h1.Set("Content-Type", "text/plain")
	m1, err := message.New(h1, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}

	h2 := make(message.Header)
	h2.Set("Content-Type", "text/html")
	m2, err := message.New(h2, bytes.NewReader(buf2.Bytes()))
	if err != nil {
		return err
	}
	to := &mail.Address{
		Name:    user.Profile.Name,
		Address: user.Email,
	}
	from, err := mail.ParseAddressList(b.Sender)
	if err != nil {
		return err
	}

	sender := mailer.FormatAddressList(from)
	h := make(message.Header)
	h.Set("Sender", sender)
	h.Set("From", sender)
	h.Set("To", mailer.FormatAddressList([]*mail.Address{to}))
	h.Set("Subject", "[Thatiq] Permintaan Reset Password")
	h.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", v1.NewObjectID().Hex()))
	msg, err := message.NewMultipart(h, []*message.Entity{m1, m2})
	if err != nil {
		return err
	}
	b.queue.Push(mailer.NewJobMail(b.transport, []*message.Entity{msg}))
	return nil
}
