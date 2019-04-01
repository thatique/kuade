package mongo

import (
	"context"

	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/storage"
	"github.com/thatique/kuade/kuade/storage/factory"
	"github.com/thatique/kuade/kuade/storage/mongo/db"
	"github.com/thatique/kuade/kuade/storage/mongo/users"
)

type MongoParam struct {
	URL string
}

type Driver struct {
	c *db.Client
}

func init() {
	factory.RegisterFunc("mongodb", func(parameters map[string]interface{}) (storage.Driver, error) {
		return FromParameters(parameters)
	})
}

func FromParameters(parameters map[string]interface{}) (*Driver, error) {
	defaultURL := "mongodb://localhost:27017/kuade"
	if v, ok := parameters["url"]; ok {
		defaultURL = v.(string)
	}

	return New(&MongoParam{URL: defaultURL})
}

func New(param *MongoParam) (*Driver, error) {
	c, err := db.Connect(context.Background(), param.URL)
	if err != nil {
		return nil, err
	}
	return &Driver{c: c}, nil
}

func (driver *Driver) Name() string {
	return "mongodb"
}

func (driver *Driver) GetUserStorage() (store auth.UserStore, err error) {
	store = users.New(driver.c)
	return
}
