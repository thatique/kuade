package iam

import (
	"crypto/rand"
	"fmt"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Enabled      bool      `xml:"Enabled" json:"enabled"`
	Password     []byte    `xml:"-" json:"-"`
	CreatedAt    time.Time `xml:"CreatedAt" json:"createdAt"`
	LastSignin   time.Time `xml:"LastSignin" json:"lastSignin"`
	RequireReset bool      `xml:"-" json:"-"`
}

func (creds Credentials) VerifyPassword(pswd []byte) bool {
	if err := bcrypt.CompareHashAndPassword(creds.Password, pswd); err != nil || !creds.Enabled {
		return false
	}

	return true
}

func GenerateCredentialsPassword(size int) ([]byte, error) {
	data := make([]byte, size)
	if n, err := rand.Read(data); err != nil {
		return nil, err
	} else if n != size {
		return nil, fmt.Errorf("Not enough data. Expected to read: %v bytes, got: %v bytes", size, n)
	}
	return data, nil
}
