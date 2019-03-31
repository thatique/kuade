package mailer

import (
	"context"

	"github.com/emersion/go-message"
	"github.com/thatique/kuade/pkg/mailer/driver"
)

type JobMail struct {
	t        driver.Transport
	messages []*message.Entity
}

func NewJobMail(t driver.Transport, messages []*message.Entity) *JobMail {
	return &JobMail{t: t, messages: messages}
}

func (j *JobMail) GetName() string {
	return "JobMail"
}

func (j *JobMail) Fire() error {
	ctx := context.Background()
	err := j.t.Open(ctx)
	if err != nil {
		return err
	}

	defer j.t.Close(ctx)

	_, err = j.t.SendMessages(ctx, j.messages)
	return err
}
