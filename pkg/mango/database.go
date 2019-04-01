package mango

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
	mu sync.Mutex
}

func (db *Database) Client() *Client {
	db.mu.Lock()
	defer db.mu.Unlock()

	cc := db.Database.Client()
	if cc == nil {
		return nil
	}
	return &Client{Client: cc}
}

func (db *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	if db.Database == nil {
		return nil
	}

	coll := db.Database.Collection(name, opts...)
	if coll == nil {
		return nil
	}
	return &Collection{Collection: coll}
}

func (db *Database) Drop(ctx context.Context) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Database.Drop", modulestr))
	defer span.end(ctx)

	err := db.Database.Drop(ctx)
	if err != nil {
		span.setError(err)
	}
	return err
}

func (db *Database) ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (*mongo.Cursor, error) {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Database.ListCollections", modulestr))
	defer span.end(ctx)

	cursor, err := db.Database.ListCollections(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return cursor, err
}

func (db *Database) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) *mongo.SingleResult {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Database.RunCommand", modulestr))
	defer span.end(ctx)

	sr := db.Database.RunCommand(ctx, runCommand, opts...)
	return sr
}

func (db *Database) RunCommandCursor(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (*mongo.Cursor, error) {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Database.RunCommandCursor", modulestr))
	defer span.end(ctx)

	cursor, err := db.Database.RunCommandCursor(ctx, runCommand, opts...)
	if err != nil {
		span.setError(err)
	}

	return cursor, err
}
