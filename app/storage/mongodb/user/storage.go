package user

import (
	"context"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage/driver"
	"github.com/thatique/kuade/app/storage/mongodb/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userStorage struct {
	c *core.Client
}

// New create userStore with mongodb backend
func New(c *core.Client) driver.UserStore {
	return &userStorage{c: c}
}

func (s *userStorage) PutUser(ctx context.Context, user *model.User) error {
	usr := fromUserModel(user)
	s.c.C(usr).FindOneAndUpdate(ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": usr},
		upsertOpts(),
	)
	return nil
}

func (s *userStorage) PutUserProfile(ctx context.Context, id model.ID, profile *model.UserProfile) error {
	_, err := s.c.C(&dbUser{}).UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"profile": fromUserProfileDomain(profile)}},
	)
	return err
}

func (s *userStorage) PutUserCredential(ctx context.Context, id model.ID, cred *model.Credentials) error {
	credentials := &dbCredentials{
		Passwords:  cred.Password,
		Enabled:    cred.Enabled,
		CreatedAt:  cred.CreatedAt,
		LastSignin: cred.LastSignin,
	}
	_, err := s.c.C(&dbUser{}).UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{"credentials": credentials}})
	return err
}

func (s *userStorage) FindOrCreateUserForProvider(ctx context.Context, userdata *model.User, provider model.OauthProvider) (newuser bool, user *model.User, err error) {
	var usr *dbUser
	var userQuery = bson.M{
		"identities": bson.M{
			"$elemMatch": bson.M{
				"name": provider.Name,
				"key":  provider.Key,
			},
		},
	}
	usr = fromUserModel(userdata)
	usr.Providers = []userProvider{
		userProvider{Name: provider.Name, Key: provider.Key},
	}

	var usr2 *dbUser
	err = s.c.C(usr).FindOneAndUpdate(ctx, userQuery, bson.M{"$setOnInsert": usr}, upsertOpts()).Decode(&usr2)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			usr.ID = usr2.ID
			return true, usr.toUserModel(), nil
		}
		return false, nil, err
	}
	return false, usr.toUserModel(), nil
}

func (s *userStorage) GetCredentialByEmail(ctx context.Context, email string) (*model.Credentials, error) {
	var usr *dbUser
	err := s.c.C(usr).FindOne(ctx,
		bson.M{"email": email},
		options.FindOne().SetProjection(bson.D{{"_id": 1}, {"credentials": 1}}),
	).Decode(&usr)
	if err != nil {
		return nil, err
	}

	return &model.Credentials{
		Email:      email,
		Password:   usr.Credentials.Passwords,
		UserID:     usr.ID,
		Enabled:    usr.Credentials.Enabled,
		CreatedAt:  usr.Credentials.CreatedAt,
		LastSignin: usr.Credentials.LastSignin,
	}, nil
}

func upsertOpts() *options.FindOneAndUpdateOptions {
	return options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.Before)
}
