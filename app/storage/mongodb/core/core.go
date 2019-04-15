package core

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/thatique/kuade/pkg/mango"
	"go.mongodb.org/mongo-driver/mongo"
)

var models = []Model{}

// Model represent mongo collection
type Model interface {
	Col() string
	Indexes() []mongo.IndexModel
}

// Client is wrapped mongo client
type Client struct {
	*mango.Client
	DB *mango.Database // default db
}

// Register a model, this will allow us to ensure index exists during startup
func Register(m Model) {
	models = append(models, m)
}

// Connect connect to mongodb server using provided url string
func Connect(ctx context.Context, urlstr string) (*Client, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %s. fail with error: %v", urlstr, err)
	}
	db := strings.Trim(u.Path, "/")

	mclient, err := mango.NewClient(urlstr)
	if err != nil {
		return nil, err
	}
	err = mclient.Connect(ctx)
	if err != nil {
		return nil, err
	}

	conn := &Client{
		Client: mclient,
		DB:     mclient.Database(db),
	}

	for _, model := range models {
		registerIndexes(conn, model)
	}

	return conn, err
}

// C return a collection for given model
func (c *Client) C(m Model) *mango.Collection {
	return c.DB.Collection(m.Col())
}

func registerIndexes(client *Client, m Model) error {
	collection := client.DB.Collection(m.Col())
	ixView := collection.Indexes()
	// then register this index
	_, err := ixView.CreateMany(context.Background(), m.Indexes())
	return err
}
