package users

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/storage/mongo/db"
	"github.com/thatique/kuade/pkg/emailparser"
)

func init() {
	db.Register(&userMgo{})
}

type dbProfile struct {
	Name    string `bson:"name,omitempty"`
	Picture string `bson:"picture,omitempty"`
	Bio     string `bson:"bio,omitempty"`

	Age     int32  `bson:"age,omitempty"`
	Address string `bson:"address,omitempty"`
	City    string `bson:"city,omitempty"`
	State   string `bson:"state,omitempty"`
}

type userProvider struct {
	Name string `bson:"name"`
	Key  string `bson:"key"`
}

type userMgo struct {
	Id        bson.ObjectId   `bson:"_id,omitempty"`
	Slug      string          `bson:"slug"`
	Profile   dbProfile       `bson:"profile,omitempty"`
	Email     string          `bson:"email"`
	Password  []byte          `bson:"password"`
	Status    auth.UserStatus `bson:"status"`
	Role      auth.Role       `bson:"role"`
	CreatedAt time.Time       `bson:"created_at"`
	Providers []userProvider  `bson:"identities"`
}

func (user *userMgo) Col() string {
	return "users"
}

func (user *userMgo) Indexes() []mgo.Index {
	return []mgo.Index{
		mgo.Index{Key: []string{"email"}, Unique: true},
		mgo.Index{Key: []string{"slug"}, Unique: true},
		mgo.Index{Key: []string{"identities.name", "identities.key"}, Unique: true, Sparse: true},
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

func (u *userMgo) Presave(conn *db.Conn) {
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
		Email:     user.Email,
		Password:  user.Password,
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
		Email:     user.Email,
		Password:  user.Password,
		Status:    user.Status,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}
