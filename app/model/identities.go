package model

// IsActive check if the user status currently active
func (m *User) IsActive() bool {
	return m.GetStatus() == UserStatus_ACTIVE
}

func (m *User) GetUserName() string {
	return m.GetEmail()
}

func (m *User) GetUID() string {
	return m.GetUID()
}

func (m *User) GetGroups() []string {
	return []string{m.GetRole().String()}
}

func (m *User) GetMetadata() map[string][]string {
	metadata := make(map[string][]string)
	return metadata
}
