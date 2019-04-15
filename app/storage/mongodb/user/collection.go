package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/storage/mongodb/core"
)

type dbProfile struct {
	Icon    string      `bson:"icon,omitempty"`
	Bio     string      `bson:"bio,omitempty"`
	Age     int32       `bson:"age,omitempty"`
	Address string      `bson:"address"`
	City    string      `bson:"city"`
	State   string      `bson:"state"`
	Zipcode string      `bson:"zipcode"`
	Point   *core.Point `bson:"point,omitempty"`
}

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
	ID        model.ID         `bson:"_id"`
	Slug      string           `bson:"slug"`
	Name      string           `bson:"name"`
	Email     string           `bson:"email"`
	Role      model.UserRole   `bson:"role"`
	Status    model.UserStatus `bson:"status"`
	CreatedAt time.Time        `bson:"created_at"`

	// embed the relation
	Profile     dbProfile      `bson:"profile,omitempty"`
	Credentials dbCredentials  `bson:"credentials,omitempty"`
	Providers   []userProvider `bson:"identities,omitempty"`
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
			Keys:    bson.D{{"role": 1}, {"status": 1}, {"created_at": -1}},
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
		ID:        user.ID,
		Slug:      user.Slug,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}

func fromUserModel(user *model.User) *dbUser {
	return &dbUser{
		ID:        user.ID,
		Slug:      user.GetSlug(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		Role:      user.GetRole(),
		Status:    user.GetStatus(),
		CreatedAt: user.GetCreatedAt(),
	}
}

func fromUserProfileDomain(profile *model.UserProfile) *dbProfile {
	p := &dbProfile{
		Icon:    profile.GetIcon(),
		Bio:     profile.GetBio(),
		Age:     profile.GetAge(),
		Address: profile.GetAddress(),
		City:    profile.GetCity(),
		State:   profile.GetState(),
		Zipcode: profile.GetZipcode(),
	}
	if point := profile.GetPoint(); point != nil {
		p.Point = &core.Point{
			Type:        "Point",
			Coordinates: [2]float64{point.GetLongitude(), point.GetLatitude()},
		}
	}
	return p
}
