package email

import (
	"io"
	"net/mail"

	"github.com/emersion/go-message"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/pkg/mailer"
	"github.com/thatique/kuade/pkg/queue"
)

type MailNotifier struct {
	Sender    string
	transport mailer.Transport
	channel   chan<- queue.Job
}

func NewMailNotifier(sender string, m mailer.Transport, c chan<- queue.Job) *MailNotifier {
	return &MailNotifier{Sender: sender, transport: m, channel: c}
}

func (n *MailNotifier) Notify(user *auth.User, body io.Reader) error {
	to := &mail.Address{
		Name:    user.Profile.Name,
		Address: user.Email,
	}
	from, err := mail.ParseAddressList(n.Sender)
	if err != nil {
		return err
	}

	sender := mailer.FormatAddressList(from)
	h := make(message.Header)
	h.Set("Sender", sender)
	h.Set("From", sender)
	h.Set("To", mailer.FormatAddressList([]*mail.Address{to}))
	h.Set("Subject", "[Thatiq] Permintaan Reset Password")
	h.Set("Content-Type", "text/plain")

	msg, err := message.New(h, body)
	if err != nil {
		return err
	}

	n.channel <- mailer.NewJobMail(n.transport, []*message.Entity{msg})
	return nil
}
