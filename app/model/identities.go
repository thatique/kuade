package model

import (
	"github.com/thatique/kuade/pkg/iam/auth/user"
)

// IsActive check if the user status currently active
func (m *User) IsActive() bool {
	return m.GetStatus() == UserStatus_ACTIVE
}

// ToAuthInfo convert user model to user.Info
func (m *User) ToAuthInfo() user.Info {
	return &user.DefaultInfo{
		Name:   m.GetEmail(),
		UID:    m.ID.String(),
		Groups: []string{m.GetRole().String()},
	}
}
