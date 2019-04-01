package users

import (
	"time"

	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/storage/mongo/db"
	"github.com/thatique/kuade/pkg/emailparser"
	"go.mongodb.org/mongo-driver/bson"
	bsonp "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	db.Register(&userMgo{})
}

type dbProfile struct {
	Name    string `bson:"name,omitempty"`
	Picture string `bson:"picture,omitempty"`
	Bio     string `bson:"bio,omitempty"`

	Age     uint8  `bson:"age,omitempty"`
	Address string `bson:"address,omitempty"`
	City    string `bson:"city,omitempty"`
	State   string `bson:"state,omitempty"`
}

type dbCredentials struct {
	Enabled    bool      `bson:"enabled"`
	Password   []byte    `bson:"password"`
	CreatedAt  time.Time `bson:"createdAt"`
	LastSignin time.Time `bson:"lastSignin"`
}

type userProvider struct {
	Name string `bson:"name"`
	Key  string `bson:"key"`
}

type userMgo struct {
	Id          bsonp.ObjectID  `bson:"_id,omitempty"`
	Slug        string          `bson:"slug"`
	Profile     dbProfile       `bson:"profile,omitempty"`
	Email       string          `bson:"email"`
	Credentials dbCredentials   `bson:"credentials"`
	Status      auth.UserStatus `bson:"status"`
	Role        auth.Role       `bson:"role"`
	CreatedAt   time.Time       `bson:"created_at"`
	Providers   []userProvider  `bson:"identities,omitempty"`
}

func (user *userMgo) Col() string {
	return "users"
}

func (user *userMgo) Indexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		mongo.IndexModel{
			Keys:    bson.D{{"email", "1"}},
			Options: (&options.IndexOptions{}).SetUnique(true),
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

func (u *userMgo) SortBy() string {
	return "-created_at"
}

func (u *userMgo) Unique() bson.M {
	if len(u.Id) > 0 {
		return bson.M{"_id": u.Id}
	}

	return bson.M{"email": u.Email}
}

func (u *userMgo) SlugQuery(slug string) bson.M {
	return bson.M{"slug": slug}
}

func (u *userMgo) Presave(conn *db.Client) {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}

	if u.Slug == "" {
		if u.Profile.Name != "" {
			slug, err := conn.GenerateSlug(u, u.Profile.Name)
			if err != nil {
				panic(err)
			}
			u.Slug = slug
		} else {
			email, _ := emailparser.NewEmail(u.Email)
			slug, err := conn.GenerateSlug(u, email.Local())
			if err != nil {
				panic(err)
			}
			u.Slug = slug
		}
	}

	if !u.Status.IsValid() {
		u.Status = auth.USER_STATUS_ACTIVE
	}
}

func fromAuthModel(user *auth.User) *userMgo {
	return &userMgo{
		Id:   user.Id,
		Slug: user.Slug,
		Profile: dbProfile{
			Name:    user.Profile.Name,
			Picture: user.Profile.Picture,
			Bio:     user.Profile.Bio,
			Age:     user.Profile.Age,
			Address: user.Profile.Address,
			City:    user.Profile.City,
			State:   user.Profile.State,
		},
		Email: user.Email,
		Credentials: dbCredentials{
			Enabled:    user.Credentials.Enabled,
			Password:   user.Credentials.Password,
			CreatedAt:  user.Credentials.CreatedAt,
			LastSignin: user.Credentials.LastSignin,
		},
		Status:    user.Status,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func toAuthModel(user *userMgo) *auth.User {
	return &auth.User{
		Id:   user.Id,
		Slug: user.Slug,
		Profile: auth.Profile{
			Name:    user.Profile.Name,
			Picture: user.Profile.Picture,
			Bio:     user.Profile.Bio,
			Age:     user.Profile.Age,
			Address: user.Profile.Address,
			City:    user.Profile.City,
			State:   user.Profile.State,
		},
		Email: user.Email,
		Credentials: auth.Credentials{
			Enabled:    user.Credentials.Enabled,
			Password:   user.Credentials.Password,
			CreatedAt:  user.Credentials.CreatedAt,
			LastSignin: user.Credentials.LastSignin,
		},
		Status:    user.Status,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
