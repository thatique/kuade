package auth

import (
	"context"
	"fmt"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/globalsign/mgo/bson"
	"github.com/thatique/kuade/kuade/api/types"
)

type UserStatus int

const (
	USER_STATUS_INACTIVE UserStatus = iota
	USER_STATUS_ACTIVE
	USER_STATUS_LOCKED
)

var statusNames = map[string]UserStatus{
	"inactive":   USER_STATUS_INACTIVE,
	"active":     USER_STATUS_ACTIVE,
	"locked":     USER_STATUS_LOCKED,
}

func (st UserStatus) IsValid() bool {
	switch st {
	case USER_STATUS_INACTIVE, USER_STATUS_ACTIVE, USER_STATUS_LOCKED:
		return true
	default:
		return false
	}
}

func (st UserStatus) String() string {
	switch st {
	case USER_STATUS_INACTIVE:
		return "inactive"
	case USER_STATUS_ACTIVE:
		return "active"
	case USER_STATUS_LOCKED:
		return "locked"
	default:
		return "unknown"
	}
}

func (st UserStatus) MarshalText() ([]byte, error) {
	return []byte(st.String()), nil
}

func (st *UserStatus) UnmarshalText(s []byte) error {
	if v, ok := statusNames[string(s)]; ok {
		*st = v
		return nil
	}
	return fmt.Errorf("unknown user status %v", string(s))
}

type Role int

const (
	ROLE_INDIVIDUAL Role = iota
	ROLE_VENDOR
	ROLE_STAFF
)

var roleIDs = map[Role]string{
	ROLE_INDIVIDUAL: "individual",
	ROLE_VENDOR:     "vendor",
	ROLE_STAFF:      "staff",
}

var roleNames = map[string]Role{
	"individual": ROLE_INDIVIDUAL,
	"vendor":     ROLE_VENDOR,
	"staff":      ROLE_STAFF,
}

func (role Role) MarshalText() ([]byte, error) {
	if v, ok := roleIDs[role]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("unknown role %v", role)
}

func (role *Role) UnmarshalText(r []byte) error {
	if v, ok := roleNames[string(r)]; ok {
		*role = v
		return nil
	}
	return fmt.Errorf("unknown role %v", string(r))
}

type Profile struct {
	Name    string `json:"name,omitempty"`
	Picture string `json:"picture,omitempty"`
	Bio     string `json:"bio,omitempty"`
	
	Age     int32  `json:"age,omitempty"`
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
}

type User struct {
	Id         bson.ObjectId   `json:"id"`
	Slug       string          `json:"slug"`
	Profile    Profile         `json:"profile,omitempty"`
	Email      string          `json:"email"`
	Password   []byte          `json:"-"`
	Status     UserStatus      `son:"status"`
	Superuser  bool            `json:"is_superuser"`
	Role       Role            `json:"-"`
	CreatedAt  time.Time       `json:"created_at"`
}

type OAuthProvider struct {
	Name string
	Key  string
}

type UserStore interface {
	// find a user by it's id
	FindById(ctx context.Context, id bson.ObjectId) (*User, error)

	// find a user by it's email
	FindByEmail(ctx context.Context, email string) (*User, error)

	// find a user by slug
	FindBySlug(ctx context.Context, slug string) (*User, error)

	FindOrCreateUserForProvider(ctx context.Context, userdata *User, provider OAuthProvider) (newuser bool, user *User, err error)

	// list returns a list of users from the datastore.
	List(ctx context.Context, pagination types.PaginationArgs) ([]*User, error)

	Create(ctx context.Context, user *User) error

	// Update persists an updated user to the datastore.
	Update(context.Context, *User) error

	// Delete deletes a user from the datastore.
	Delete(context.Context, *User) error

	// Count returns a count of active users.
	Count(context.Context) (int, error)
}

func (user *User) IsActive() bool {
	return user.Status == USER_STATUS_ACTIVE
}

func (user *User) SetPassword(pswd []byte) error {
	b, err := bcrypt.GenerateFromPassword(pswd, 11)
	if err != nil {
		return err
	}
	user.Password = b
	return nil
}

func (user *User) VerifyPassword(pswd []byte) bool {
	if err := bcrypt.CompareHashAndPassword(user.Password, pswd); err != nil {
		return false
	}

	return true 
}