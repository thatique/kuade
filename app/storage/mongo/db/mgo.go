package db

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/thatique/kuade/pkg/mango"
	"github.com/thatique/kuade/pkg/text"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var models = []Model{}

// A model represent mongo collection
type Model interface {
	Col() string
	Indexes() []mongo.IndexModel
}

type Slugable interface {
	Model
	SlugQuery(slug string) bson.M
}

type OrderedModel interface {
	Model
	SortBy() string
}

type Updatable interface {
	Model
	Unique() bson.M
	Presave(client *Client)
}

type Client struct {
	*mango.Client
	DB *mango.Database // default db
}

func Register(m Model) {
	models = append(models, m)
}

func registerIndexes(client *Client, m Model) error {
	collection := client.DB.Collection(m.Col())
	ixView := collection.Indexes()
	// then register this index
	_, err := ixView.CreateMany(context.Background(), m.Indexes())
	return err
}

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

//
func (c *Client) C(m Model) *mango.Collection {
	return c.DB.Collection(m.Col())
}

func (c *Client) GenerateSlug(m Slugable, base string) (string, error) {
	var (
		slug       = text.Slugify(base)
		collection = c.C(m)
		maxretries = 20
		retries    int
		err        error
		ret        bson.M
	)
	slugToTry := slug
	for {
		err = collection.FindOne(context.Background(), m.SlugQuery(slugToTry)).Decode(&ret)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return slugToTry, nil
			}
			// non expected error, return
			return "", err
		}
		retries += 1
		if retries > maxretries {
			return "", fmt.Errorf("generateslug: maximum retries reached. max: %d", maxretries)
		}
		slugToTry = fmt.Sprintf("%s-%d", slug, retries)
	}
}
