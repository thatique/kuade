package users

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	bsonp "go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/storage"
	"github.com/thatique/kuade/kuade/storage/mongo/db"
)

type MgoUserStore struct {
	c *db.Client
}

func New(conn *db.Client) *MgoUserStore {
	return &MgoUserStore{c: conn}
}

func (conn *MgoUserStore) FindById(ctx context.Context, id bsonp.ObjectID) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.C(dbuser).FindOne(ctx, bson.M{"_id": id}).Decode(&dbuser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = storage.ErrNotFound
		}
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindByEmail(ctx context.Context, email string) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.C(dbuser).FindOne(ctx, bson.M{"email": email}).Decode(&dbuser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = storage.ErrNotFound
		}
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindBySlug(ctx context.Context, slug string) (user *auth.User, err error) {
	var dbuser *userMgo
	err = conn.c.C(dbuser).FindOne(ctx, bson.M{"slug": slug}).Decode(&dbuser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
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

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.Before)

	var dbuser2 *userMgo
	err = conn.c.C(dbuser).FindOneAndUpdate(ctx, userQuery, bson.M{"$setOnInsert": dbuser}, opts).Decode(&dbuser2)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			dbuser.Id = dbuser2.Id
			return true, toAuthModel(dbuser), nil
		}
		return false, nil, err
	}

	return false, toAuthModel(dbuser), nil
}

func (conn *MgoUserStore) Create(ctx context.Context, user *auth.User) (id bsonp.ObjectID, err error) {
	dbuser := fromAuthModel(user)
	dbuser.Presave(conn.c)
	info, err := conn.c.C(dbuser).InsertOne(ctx, dbuser)

	id, ok := info.InsertedID.(bsonp.ObjectID)
	if !ok {
		return bsonp.NilObjectID, err
	}
	return id, nil
}

func (conn *MgoUserStore) Update(ctx context.Context, user *auth.User) (err error) {
	if len(user.Id) == 0 {
		err = errors.New("storage mongo: user didn't exist")
		return
	}

	dbuser := fromAuthModel(user)
	dbuser.Presave(conn.c)
	_, err = conn.c.C(dbuser).UpdateOne(ctx,  bson.M{"_id": user.Id}, bson.M{"$set": dbuser})
	return
}

func (conn *MgoUserStore) UpdateCredentials(ctx context.Context, id bsonp.ObjectID, creds auth.Credentials) error {
	var dbCreds = dbCredentials{
		Enabled:    creds.Enabled,
		Password:   creds.Password,
		CreatedAt:  creds.CreatedAt,
		LastSignin: creds.LastSignin,
	}
	_, err := conn.c.C(&userMgo{}).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"credentials": dbCreds})
	return err
}

func (conn *MgoUserStore) Delete(ctx context.Context, user *auth.User) (err error) {
	if len(user.Id) == 0 {
		err = errors.New("storage mongo: user didn't exist")
		return
	}

	_, err = conn.c.C(&userMgo{}).DeleteOne(ctx, bson.M{"_id": user.Id})
	return
}

func (conn *MgoUserStore) Count(ctx context.Context) (count int64, err error) {
	count, err = conn.c.C(&userMgo{}).CountDocuments(ctx, bson.M{})
	return
}
