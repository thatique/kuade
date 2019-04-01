package mango

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type Session struct {
	mongo.Session
}

var _ mongo.Session = (*Session)(nil)

func (sess *Session) EndSession(ctx context.Context) {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Session.EndSession", modulestr))
	defer span.end(ctx)

	sess.Session.EndSession(ctx)
}

func (sess *Session) AbortTransaction(ctx context.Context) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Session.AbortTransaction", modulestr))
	defer span.end(ctx)

	err := sess.Session.AbortTransaction(ctx)
	if err != nil {
		span.setError(err)
	}
	return err
}

func (sess *Session) CommitTransaction(ctx context.Context) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Session.CommitTransaction", modulestr))
	defer span.end(ctx)

	err := sess.Session.CommitTransaction(ctx)
	if err != nil {
		span.setError(err)
	}
	return err
}
