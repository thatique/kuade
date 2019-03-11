package users

import (
	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/storage/mongo"
)

type MgoUserStore struct {
	c *mongo.Conn
}

func (conn *MgoUserStore) FindById(ctx context.Context, id bson.ObjectId) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.C(dbuser).FindId(id).One(&dbuser)
	if err != nil {
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindByEmail(ctx context.Context, email string) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.Find(dbuser, bson.M{"email": email}).One(&dbuser)
	if err != nil {
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindByEmail(ctx context.Context, slug string) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.Find(dbuser, bson.M{"slug": slug}).One(&dbuser)
	if err != nil {
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindOrCreateUserForProvider(ctx context.Context, data *auth.User, provider auth.Provider) (user *auth.User, err error) {
	var dbuser *userMgo
	var userQuery = bson.M{
		"identities": bson.M{
			"$elemMatch": bson.M{
				"name": provider.Name,
				"key": provider.Key,
			},
		},
	}

	dbuser = fromAuthModel(data)
	dbuser.Providers = []userProvider{
		userProvider{Name: provider.Name, Key: provider.Key,},
	}

	info, err := conn.c.C(dbuser).Upsert(userQuery, bson.M{"$setOnInsert": dbuser})
	if err != nil {
		return
	}

	
}