package mango

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const modulestr = "go.mongodb.org/mongo-driver/mongo"

type Client struct {
	*mongo.Client
}

func NewClient(urlstr string) (*Client, error) {
	mclient, err := mongo.NewClient(options.Client().ApplyURI(urlstr))
	if err != nil {
		return nil, err
	}

	return &Client{Client: mclient}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Client.Connect", modulestr))
	defer span.end(ctx)

	err := c.Client.Connect(ctx)
	if err != nil {
		span.setError(err)
	}
	return err
}

func (c *Client) Database(name string, opts ...*options.DatabaseOptions) *Database {
	db := c.Client.Database(name, opts...)
	if db == nil {
		return nil
	}

	return &Database{Database: db}
}

func (c *Client) Disconnect(ctx context.Context) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Client.Disconnect", modulestr))
	defer span.end(ctx)

	err := c.Client.Disconnect(ctx)
	if err != nil {
		span.setError(err)
	}
	return err
}

func (c *Client) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error) {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Client.ListDatabaseNames", modulestr))
	defer span.end(ctx)

	dbs, err := c.Client.ListDatabaseNames(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return dbs, err
}

func (c *Client) ListDatabases(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (mongo.ListDatabasesResult, error) {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Client.ListDatabases", modulestr))
	defer span.end(ctx)

	dbr, err := c.Client.ListDatabases(ctx, filter, opts...)
	if err != nil {
		span.setError(err)
	}
	return dbr, err
}

func (c *Client) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	ctx, span := roundtripTrackingSpan(ctx, fmt.Sprintf("%s.Client.Ping", modulestr))
	defer span.end(ctx)

	err := c.Client.Ping(ctx, rp)
	if err != nil {
		span.setError(err)
	}
	return err
}

func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	ss, err := c.Client.StartSession(opts...)
	if err != nil {
		return nil, err
	}
	return &Session{Session: ss}, nil
}

func (c *Client) GetClient() *mongo.Client {
	return c.Client
}
