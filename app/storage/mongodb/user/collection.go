package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/thatique/kuade/app/model"
)

type dbCredentials struct {
	Passwords  []byte    `bson:"password"`
	Enabled    bool      `bson:"enabled"`
	CreatedAt  time.Time `bson:"created_at"`
	LastSignin time.Time `bson:"last_signin"`
}

type userProvider struct {
	Name string `bson:"name"`
	Key  string `bson:"key"`
}

type dbUser struct {
	ID       int64         `bson:"_id"`
	Slug     string        `bson:"slug"`
	Name     string        `bson:"name"`
	Email    string        `bson:"email"`
	Icon     string        `bson:"icon"`
	Role     int32         `bson:"role"`
	Status   int32         `bson:"status"`
	Bio      string        `bson:"bio,omitempty"`
	Age      int32         `bson:"age,omitempty"`
	Address  model.Address `bson:"address,omitempty"`
	Category string        `bson:"category,omitempty"`
	Budget   int32         `bson:"budget"`
	// embed the relation
	Credentials dbCredentials  `bson:"credentials,omitempty"`
	Providers   []userProvider `bson:"identities,omitempty"`
	CreatedAt   time.Time      `bson:"created_at"`
}

// Coll return collection name for storing dbUser
func (user *dbUser) Col() string {
	return "users"
}

func (user *dbUser) Indexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		mongo.IndexModel{
			Keys:    bson.D{{"email", "1"}},
			Options: (&options.IndexOptions{}).SetUnique(true),
		},
		mongo.IndexModel{
			Keys:    bson.D{{"role": 1}, {"status": 1}, {"category": 1}, {"budget": -1}},
			Options: &options.IndexOptions{},
		},
		mongo.IndexModel{
			Keys:    bson.D{{"slug", "1"}},
			Options: (&options.IndexOptions{}).SetUnique(true),
		},
		mongo.IndexModel{
			Keys:    bson.D{{"identities.name", 1}, {"identities.key", 1}},
			Options: (&options.IndexOptions{}).SetUnique(true).SetSparse(true),
		},
	}
}

func (user *dbUser) toUserModel() *model.User {
	return &model.User{
		ID:        model.ID(uint64(user.ID)),
		Slug:      user.Slug,
		Name:      user.Name,
		Email:     user.Email,
		Role:      model.UserRole(user.Role),
		Status:    model.UserStatus(user.Status),
		Bio:       user.Bio,
		Age:       user.Age,
		Address:   user.Address,
		Category:  user.Category,
		Budget:    model.BudgetLevel(user.Budget),
		CreatedAt: user.CreatedAt,
	}
}

func fromUserModel(user *model.User) *dbUser {
	return &dbUser{
		ID:        int64(user.ID),
		Slug:      user.GetSlug(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		Role:      int32(user.GetRole()),
		Status:    int32(user.GetStatus()),
		Bio:       user.GetBio(),
		Age:       user.GetAge(),
		Address:   user.GetAddress(),
		Category:  user.Category,
		Budget:    int32(user.Budget),
		CreatedAt: user.GetCreatedAt(),
	}
}
