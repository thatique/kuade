package passwords

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	textTemplate "text/template"
	"time"

	"github.com/emersion/go-message"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/auth/password/driver"
	"github.com/thatique/kuade/pkg/emailparser"
	"github.com/thatique/kuade/pkg/mailer"
	"github.com/thatique/kuade/pkg/uuid"
)

var ErrInvalidEmail = errors.New("passwords: invalid email")

type ErrorCode int

const (
	NoError ErrorCode = iota
	ErrPassNotMatch
	ErrMinimumPassword
	ErrUpstream
)

type Broker struct {
	ResetURL string
	Sender   string
	Mailer   *mailer.Transport
	tokens   driver.ResetTokenGenerator
	users    *storage.UserStore
	Text     *textTemplate.Template
	HTML     *template.Template
}

type Request struct {
	creds                       *model.Credentials
	token, Password1, Password2 string
}

func New(tokens driver.ResetTokenGenerator, users *storage.UserStore) *Broker {
	return &Broker{
		tokens: tokens,
		users:  users,
	}
}

func (b *Broker) SendResetLink(ctx context.Context, ip string, email string) error {
	if !emailparser.IsValidEmail(email) {
		return ErrInvalidEmail
	}

	creds, err := b.users.GetCredentialByEmail(ctx, email)
	if err != nil {
		return err
	}

	token, err := b.tokens.Create(ctx, creds)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/%x/%s", b.ResetURL, uint64(creds.UserID), token)
	tplCtx := map[string]interface{}{
		"Link":      link,
		"Email":     creds.GetEmail(),
		"CreatedAt": time.Now().UTC().Format(time.RFC1123),
		"IP":        ip,
	}

	msg := b.composeEmail(ctx, creds, tplCtx)
	_, err := b.SendMessages(ctx, []*message.Entity{msg})
	return err
}

func (b *Broker) Resets(ctx context.Context, req *Request) ErrorCode {
	if req.Password1 != req.Password2 {
		return ErrPassNotMatch
	}

	if req.Password1 < 8 {
		return ErrMinimumPassword
	}

	creds := req.creds
	creds.SetPassword([]byte(req.Password1))
	if err := b.users.PutUserCredential(ctx, creds); err != nil {
		return ErrUpstream
	}

	return NoError
}

func (b *Broker) ValidateReset(ctx context.Context, uid, token string) (req *Request, ok bool) {
	if uid == "" || token == "" {
		ok = false
		return
	}

	id, err := model.NewIDFromString(uid)
	if err != nil {
		ok = false
		return
	}

	user, err := b.users.GetUserByID(ctx, id)
	if err != nil {
		ok = false
		return
	}
	creds, err := b.users.GetCredentialByEmail(ctx, user.GetEmail())
	if err != nil {
		ok = false
		return
	}

	ok = b.tokens.Check(ctx, creds, token)
	if !ok {
		return
	}

	req = &Request{
		creds: creds,
		token: token,
	}
	ok = true
	return
}

func (b *Broker) composeEmail(ctx context.Context, creds *model.Credentials, tplCtx map[string]interface{}) *message.Entity {
	boundary := uuid.Generate()

	h1 := make(message.Header)
	h1.Set("Content-Type", "text/plain")
	var textBuf bytes.Buffer
	if err := b.Text.Execute(&textBuf, tplCtx); err != nil {
		panic(err)
	}
	e1, _ := message.New(h1, bytes.NewReader(textBuf.Bytes()))

	h2 := make(message.Header)
	h2.Set("Content-Type", "text/html")
	var htmlBuf bytes.Buffer
	if err := b.HTML.Execute(&htmlBuf, tplCtx); err != nil {
		panic(err)
	}
	e2, _ := message.New(h2, bytes.NewReader(htmlBuf.Bytes()))

	h := make(message.Header)
	h.Set("Sender", h.Sender)
	h.Set("From", h.Sender)
	h.Set("To", creds.GetEmail())
	h.Set("Subject", "[Thatique.com] Reset Password")
	h.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%", boundary.String()))
	msg, _ := message.NewMultipart(h, []*message.Entity{e1, e2})

	return msg
}
