package users

import (
	"context"
	"errors"

	"github.com/thatique/kuade/api/v1"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/app/storage/mongo/db"
	"go.mongodb.org/mongo-driver/bson"
	bsonp "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MgoUserStore struct {
	c *db.Client
}

func New(conn *db.Client) *MgoUserStore {
	return &MgoUserStore{c: conn}
}

func (conn *MgoUserStore) FindUserById(ctx context.Context, id v1.ObjectID) (user *model.User, err error) {
	var dbuser *userMgo
	err = conn.c.C(dbuser).FindOne(ctx, bson.M{"_id": db.ToMongoObjectID(id)}).Decode(&dbuser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = storage.ErrNotFound
		}
		return
	}
	user = toAuthModel(dbuser)
	return
}

func (conn *MgoUserStore) FindUserByEmail(ctx context.Context, email string) (user *model.User, err error) {
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

func (conn *MgoUserStore) FindUserBySlug(ctx context.Context, slug string) (user *model.User, err error) {
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

func (conn *MgoUserStore) FindOrCreateUserForProvider(ctx context.Context, data *model.User, provider model.OAuthProvider) (newuser bool, user *model.User, err error) {
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
			dbuser.ID = dbuser2.ID
			return true, toAuthModel(dbuser), nil
		}
		return false, nil, err
	}

	return false, toAuthModel(dbuser), nil
}

func (conn *MgoUserStore) InsertUser(ctx context.Context, user *model.User) (id v1.ObjectID, err error) {
	dbuser := fromAuthModel(user)
	dbuser.Presave(conn.c)
	info, err := conn.c.C(dbuser).InsertOne(ctx, dbuser)

	oid, ok := info.InsertedID.(bsonp.ObjectID)
	if !ok {
		return v1.NilObjectID, err
	}
	id = db.FromMongoObjectID(oid)
	return id, nil
}

func (conn *MgoUserStore) UpdateUser(ctx context.Context, user *model.User) (err error) {
	if len(user.ID) == 0 {
		err = errors.New("storage mongo: user didn't exist")
		return
	}

	dbuser := fromAuthModel(user)
	dbuser.Presave(conn.c)
	_, err = conn.c.C(dbuser).UpdateOne(ctx, dbuser.Unique(), bson.M{"$set": dbuser})
	return
}

func (conn *MgoUserStore) UpdateUserCredentials(ctx context.Context, id v1.ObjectID, creds model.Credentials) error {
	var dbCreds = dbCredentials{
		Enabled:    creds.Enabled,
		Password:   creds.Password,
		CreatedAt:  creds.CreatedAt,
		LastSignin: creds.LastSignin,
	}
	_, err := conn.c.C(&userMgo{}).UpdateOne(ctx, bson.M{"_id": db.ToMongoObjectID(id)},
		bson.M{"$set": bson.M{"credentials": dbCreds}})
	return err
}

func (conn *MgoUserStore) UpdateUserProfile(ctx context.Context, id v1.ObjectID, profile model.Profile) error {
	var prof = dbProfile{
		Name:    profile.Name,
		Picture: profile.Picture,
		Bio:     profile.Bio,
		Age:     profile.Age,
		Address: profile.Address,
		City:    profile.City,
		State:   profile.State,
	}
	_, err := conn.c.C(&userMgo{}).UpdateOne(ctx, bson.M{"_id": db.ToMongoObjectID(id)},
		bson.M{"$set": bson.M{"profile": prof}})
	return err
}

func (conn *MgoUserStore) DeleteUser(ctx context.Context, user *model.User) (err error) {
	if len(user.ID) == 0 {
		err = errors.New("storage mongo: user didn't exist")
		return
	}

	_, err = conn.c.C(&userMgo{}).DeleteOne(ctx, bson.M{"_id": db.ToMongoObjectID(user.ID)})
	return
}
