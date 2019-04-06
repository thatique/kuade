package model

import (
	"time"

	"github.com/thatique/kuade/api/v1"
	"github.com/thatique/kuade/pkg/auth/user"
	"golang.org/x/crypto/bcrypt"
)

type Profile struct {
	Name    string `json:"name,omitempty"`
	Picture string `json:"picture,omitempty"`
	Bio     string `json:"bio,omitempty"`

	Age     uint8  `json:"age,omitempty"`
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
}

type Credentials struct {
	Enabled    bool      `xml:"Enabled" json:"enabled"`
	Password   []byte    `xml:"-" json:"-"`
	CreatedAt  time.Time `xml:"CreatedAt" json:"createdAt"`
	LastSignin time.Time `xml:"LastSignin" json:"lastSignin"`
}

type OAuthProvider struct {
	Name string
	Key  string
}

type User struct {
	ID          v1.ObjectID `xml:"ID" json:"id"`
	Slug        string      `xml:"Slug" json:"slug"`
	Profile     Profile     `xml:"Profile" json:"profile,omitempty"`
	Email       string      `xml:"-" json:"-"`
	Credentials Credentials `xml:"-" json:"-"`
	Status      UserStatus  `xml:"Status" json:"status"`
	Role        UserRole    `xml:"Role" json:"role"`
	CreatedAt   time.Time   `xml:"CreatedAt" json:"created_at"`
}

func (u *User) IsActive() bool {
	return u.Status == UserStatus_ACTIVE
}

func (u *User) SetPassword(pswd []byte) error {
	b, err := bcrypt.GenerateFromPassword(pswd, 11)
	if err != nil {
		return err
	}
	u.Credentials.Password = b
	return nil
}

func (u *User) VerifyPassword(pswd []byte) bool {
	if !u.Credentials.Enabled {
		bcrypt.GenerateFromPassword(pswd, 11)
		return false
	}
	if err := bcrypt.CompareHashAndPassword(u.Credentials.Password, pswd); err != nil {
		return false
	}

	return true
}

func (u *User) ToAuthInfo() user.Info {
	return &user.DefaultInfo{
		Name:     u.Email,
		UID:      u.ID.Hex(),
		Groups:   []string{u.Role.String()},
		Metadata: make(map[string][]string),
	}
}
