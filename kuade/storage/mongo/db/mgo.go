package db

import (
	"context"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/thatique/kuade/pkg/text"
)

var models = []Model{}

// A model represent mongo collection
type Model interface {
	Col()     string
	Indexes() []mgo.Index
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
	Presave(conn *Conn)
}

type Conn struct {
	Session *mgo.Session
	DB      *mgo.Database // default db
}

func Register(m Model) {
	models = append(models, m)
}

func registerIndexes(conn *Conn, m Model) error {
	collection := conn.DB.C(m.Col())
	indexes := m.Indexes()
	for _, index := range indexes {
		err := collection.EnsureIndex(index)
		if err != nil {
			return err
		}
	}
	return nil
}

func DialWithInfo(info *mgo.DialInfo) (*Conn, error) {
	session, err := mgo.DialWithInfo(info)

	conn := &Conn{
		Session: session,
		DB:      session.DB(info.Database),
	}

	for _, model := range models {
		registerIndexes(conn, model)
	}

	return conn, err
}

func (conn *Conn) Copy() *Conn {
	sess := conn.Session.Copy()
	return &Conn{
		Session: sess,
		DB:      sess.DB(conn.DB.Name),
	}
}

func (conn *Conn) Close() {
	conn.Session.Close()
}

//
func (conn *Conn) C(m Model) *mgo.Collection {
	return conn.DB.C(m.Col())
}

func (conn *Conn) Find(m Model, query interface{}) *mgo.Query {
	return conn.C(m).Find(query)
}

func (conn *Conn) Latest(ord OrderedModel, query interface{}) *mgo.Query {
	return conn.Find(ord.(Model), query).Sort(ord.SortBy())
}

func (conn *Conn) Exists(u Updatable) bool {
	var data interface{}
	err := conn.C(u.(Model)).Find(u.Unique()).One(&data)
	if err != nil {
		return false
	}
	return true
}

func (conn *Conn) Upsert(u Updatable) (info *mgo.ChangeInfo, err error) {
	u.Presave(conn)
	return conn.C(u.(Model)).Upsert(u.Unique(), u)
}

func (conn *Conn) GenerateSlug(m Slugable, base string) (string, error) {
	var (
		slug       = text.Slugify(base)
		collection = conn.DB.C(m.Col())
		maxretries = 20
		retries    int
		count      int
		err        error
	)
	slugToTry := slug
	for {
		count, err = collection.Find(m.SlugQuery(slugToTry)).Count()
		if err != nil {
			return "", err
		}
		if count == 0 {
			return slugToTry, nil
		}
		retries += 1
		if retries > maxretries {
			return "", fmt.Errorf("generateslug: maximum retries reached. max: %d", maxretries)
		}
		slugToTry = fmt.Sprintf("%s-%d", slug, retries)
	}
}

//
func (conn *Conn) WithContext(ctx context.Context, f func(*Conn) error) error {
	sess := conn.Copy()
	defer sess.Close()

	c := make(chan error, 1)
	go func() {
		c <- f(sess)
	}()

	select {
	case <-ctx.Done():
		<-c // Wait for f to return
		return ctx.Err()
	case err := <-c:
		return err
	}
}
