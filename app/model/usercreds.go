package model

import (
	"golang.org/x/crypto/bcrypt"
)

// SetPassword hash password credential
func (creds *Credentials) SetPassword(pswd []byte) error {
	b, err := bcrypt.GenerateFromPassword(pswd, 11)
	if err != nil {
		return err
	}
	creds.Password = b
	return err
}

// VerifyPassword verify the given password
func (creds *Credentials) VerifyPassword(pswd []byte) bool {
	if !creds.Enabled {
		bcrypt.GenerateFromPassword(pswd, 11)
		return false
	}

	if err := bcrypt.CompareHashAndPassword(u.Password, pswd); err != nil {
		return false
	}

	return true
}
