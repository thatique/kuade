package users

import (
	"context"
	"errors"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/thatique/kuade/kuade/api/types"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/storage"
	"github.com/thatique/kuade/kuade/storage/mongo/db"
)

type MgoUserStore struct {
	c *db.Conn
}

func New(conn *db.Conn) *MgoUserStore {
	return &MgoUserStore{c: conn}
}

func (conn *MgoUserStore) FindById(ctx context.Context, id bson.ObjectId) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.C(dbuser).FindId(id).One(&dbuser)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindByEmail(ctx context.Context, email string) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.Find(dbuser, bson.M{"email": email}).One(&dbuser)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindBySlug(ctx context.Context, slug string) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.Find(dbuser, bson.M{"slug": slug}).One(&dbuser)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindOrCreateUserForProvider(ctx context.Context, data *auth.User, provider auth.OAuthProvider) (newuser bool, user *auth.User, err error) {
	var dbuser *userMgo
	var userQuery = bson.M{
		"identities": bson.M{
			"$elemMatch": bson.M{
				"name": provider.Name,
				"key":  provider.Key,
			},
		},
	}

	dbuser = fromAuthModel(data)
	dbuser.Providers = []userProvider{
		userProvider{Name: provider.Name, Key: provider.Key},
	}

	info, err := conn.c.C(dbuser).Upsert(userQuery, bson.M{"$setOnInsert": dbuser})
	if err != nil {
		return
	}

	if info.UpsertedId != nil {
		newuser = true
		user, err = conn.FindById(context.Background(), info.UpsertedId.(bson.ObjectId))
		if err != nil {
			return
		}
	} else {
		newuser = false
		var dbuser2 *userMgo
		err = conn.c.Find(dbuser, userQuery).One(&dbuser2)
		if err != nil {
			return
		}
		user = toAuthModel(dbuser2)
	}

	return
}

func (conn *MgoUserStore) List(ctx context.Context, pagination types.PaginationArgs) (users []*auth.User, err error) {
	var dbuser *userMgo
	iter := conn.c.Latest(dbuser, nil).Skip(pagination.Offset).Limit(pagination.Limit).Iter()

	for iter.Next(&dbuser) {
		users = append(users, toAuthModel(dbuser))
	}
	if iter.Timeout() {
		err = errors.New("storage mongo: iter timeout")
		return
	}
	err = iter.Close()
	return
}

func (conn *MgoUserStore) Create(ctx context.Context, user *auth.User) (id bson.ObjectId, err error) {
	dbuser := fromAuthModel(user)
	info, err := conn.c.Upsert(dbuser)

	id, ok := info.UpsertedId.(bson.ObjectId)
	if !ok {
		return bson.ObjectId(""), err
	}
	return id, nil
}

func (conn *MgoUserStore) Update(ctx context.Context, user *auth.User) (err error) {
	if len(user.Id) == 0 {
		err = errors.New("storage mongo: user didn't exist")
		return
	}

	dbuser := fromAuthModel(user)
	err = conn.c.C(dbuser).UpdateId(dbuser.Id, bson.M{"$set": dbuser})
	return
}

func (conn *MgoUserStore) UpdateCredentials(ctx context.Context, id bson.ObjectId, creds auth.Credentials) error {
	var dbCreds = dbCredentials{
		Enabled:    creds.Enabled,
		Password:   creds.Password,
		CreatedAt:  creds.CreatedAt,
		LastSignin: creds.LastSignin,
	}
	err := conn.c.C(&userMgo{}).UpdateId(id, bson.M{"$set": bson.M{"credentials": dbCreds}})
	return err
}

func (conn *MgoUserStore) Delete(ctx context.Context, user *auth.User) (err error) {
	if len(user.Id) == 0 {
		err = errors.New("storage mongo: user didn't exist")
		return
	}

	err = conn.c.C(&userMgo{}).RemoveId(user.Id)
	return
}

func (conn *MgoUserStore) Count(ctx context.Context) (count int, err error) {
	count, err = conn.c.Find(&userMgo{}, nil).Count()
	return
}
