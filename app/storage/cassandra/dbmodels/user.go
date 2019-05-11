package dbmodels

import (
	"time"

	"github.com/thatique/kuade/app/model"
)

type User struct {
	ID        int64
	Slug      string
	Name      string
	Email     string
	Username  string
	Icon      string
	Role      int32
	Status    int32
	Bio       string
	Age       int32
	Address   *Address
	Category  string
	Budget    int32
	CreatedAt time.Time
}

type Credentials struct {
	Email      string
	Username   string
	UserID     int64
	Password   []byte
	Enabled    bool
	CreatedAt  time.Time
	LastSignin time.Time
}

func FromDomainUser(user *model.User) *User {
	return &User{
		ID:        int64(user.ID),
		Slug:      user.GetSlug(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		Username:  user.GetUsername(),
		Role:      int32(user.GetRole()),
		Status:    int32(user.GetStatus()),
		Bio:       user.GetBio(),
		Age:       user.GetAge(),
		Address:   FromDomainAddress(user.GetAddress()),
		Category:  user.Category,
		Budget:    int32(user.Budget),
		CreatedAt: user.GetCreatedAt(),
	}
}

func (user *User) ToDomain() *model.User {
	return &model.User{
		ID:        model.ID(uint64(user.ID)),
		Slug:      user.Slug,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		Role:      model.UserRole(user.Role),
		Status:    model.UserStatus(user.Status),
		Bio:       user.Bio,
		Age:       user.Age,
		Address:   user.Address.ToDomain(),
		Category:  user.Category,
		Budget:    model.BudgetLevel(user.Budget),
		CreatedAt: user.CreatedAt,
	}
}

func FromDomainUserCredential(creds *model.Credentials) *Credentials {
	return &Credentials{
		UserID:     int64(creds.UserID),
		Email:      creds.GetEmail(),
		Username:   creds.GetUsername(),
		Password:   creds.GetPassword(),
		Enabled:    creds.GetEnabled(),
		CreatedAt:  creds.GetCreatedAt(),
		LastSignin: creds.GetLastSignin(),
	}
}

func (creds *Credentials) ToDomain() *model.Credentials {
	return &model.Credentials{
		UserID:     model.ID(creds.UserID),
		Email:      creds.Email,
		Username:   creds.Username,
		Password:   creds.Password,
		Enabled:    creds.Enabled,
		CreatedAt:  creds.CreatedAt,
		LastSignin: creds.LastSignin,
	}
}
