package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
	"github.com/thatique/kuade/kuade/mongo"
	"github.com/thatique/kuade/pkg/emailparser"
)

type UserStatus string

const (
	USER_STATUS_INACTIVE UserStatus = "inactive"
	USER_STATUS_ACTIVE   UserStatus = "active"
	USER_STATUS_LOCKED   UserStatus = "locked"
)

func (st UserStatus) IsValid() bool {
	switch st {
	case USER_STATUS_INACTIVE, USER_STATUS_ACTIVE, USER_STATUS_LOCKED:
		return true
	default:
		return false
	}
}

func (st UserStatus) GetBSON() (interface{}, error) {
	if !st.IsValid() {
		return nil, fmt.Errorf("Invalid user status %s must be one of [inactive, active, locked]", st)
	}
	return string(st), nil
}

func (st *UserStatus) SetBSON(raw bson.Raw) error {
	var sts string
	err := raw.Unmarshal(&sts)
	if err != nil {
		return err
	}

	status := UserStatus(strings.ToLower(sts))
	if !status.IsValid() {
		return fmt.Errorf("Invalid user status %v must be one of [inactive, active, locked]", st)
	}

	*st = status
	return nil
}

type UserType int

const (
	USER_TYPE_INDIVIDUAL UserType = iota
	USER_TYPE_VENDOR
	USER_TYPE_STAFF
)

type Profile struct {
	Name    string `bson:"name,omitempty" json:"name,omitempty"`
	Picture string `bson:"picture,omitempty" json:"picture,omitempty"`
	Bio     string `bson:"bio,omitempty" json:"bio,omitempty"`
	
	Age     int32  `bson:"age,omitempty" json:"age,omitempty"`
	Address string `bson:"address,omitempty" json:"address,omitempty"`
	City    string `bson:"city,omitempty" json:"city,omitempty"`
	State   string `bson:"state,omitempty" json:"state,omitempty"`
}

type User struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	Slug       string          `bson:"slug"`
	Profile    Profile         `bson:"profile,omitempty" json:"profile,omitempty"`
	Email      string          `bson:"email" json:"email"`
	Password   []byte          `bson:"password" json:"-"`
	Status     UserStatus      `bson:"status" json:"status"`
	Superuser  bool            `bson:"is_superuser" json:"is_superuser"`
	Type       UserType        `bson:"type" json:"-"`
	CreatedAt  time.Time       `bson:"created_at" json:"created_at"`
	Identities []OAuthProvider `bson:"identities,omitempty" json:"-"`
}

type UserStore interface {
	// find a user by it's id
	FindById(ctx context.Context, id bson.ObjectId) (*User, error)

	// find a user by it's email
	FindByEmail(ctx context.Context, email string) (*User, error)

	// find a user by slug
	FindBySlug(ctx context.Context, slug string) (*User, error)

	// find a user by oauth provider
	FindByProvider(ctx context.Context, provider OAuthProvider) (*User, error)

	// list returns a list of users from the datastore.
	List(ctx context.Context) ([]*User, error)

	Create(ctx context.Context, user *User) error

	// Update persists an updated user to the datastore.
	Update(context.Context, *User) error

	// Delete deletes a user from the datastore.
	Delete(context.Context, *User) error

	// Count returns a count of active users.
	Count(context.Context) (int64, error)
}

type OAuthProvider struct {
	Name string `json:"name" bson:"name"`
	Key  string `json:"-" bson:"key"`
}

func (user *User) Col() string {
	return "users"
}

func (user *User) Indexes() []mgo.Index {
	return []mgo.Index{
		mgo.Index{Key: []string{"email"}, Unique: true},
		mgo.Index{Key: []string{"slug"}, Unique: true},
		mgo.Index{Key: []string{"identities.name", "identities.key"}, Unique: true, Sparse: true},
	}
}

func (u *User) SortBy() string {
	return "-created_at"
}

func (u *User) Unique() bson.M {
	if len(u.Id) > 0 {
		return bson.M{"_id": u.Id}
	}

	return bson.M{"email": u.Email}
}

func (u *User) SlugQuery(slug string) bson.M {
	return bson.M{"slug": slug}
}

func (u *User) Presave(conn *mongo.Conn) {
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

	if len(u.Status) == 0 {
		u.Status = USER_STATUS_ACTIVE
	}
}

func (user *User) SetPassword(pswd []byte) error {
	b, err := bcrypt.GenerateFromPassword(pswd, 11)
	if err != nil {
		return err
	}
	user.Password = b
	return nil
}

func (user *User) VerifyPassword(pswd string) bool {
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(pswd)); err != nil {
		return false
	}

	return true
}

func (user *User) IsActive() bool {
	return user.Status == USER_STATUS_ACTIVE
}