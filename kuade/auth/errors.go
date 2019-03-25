package auth

import (
	"errors"
)

var (
	ErrAlreadySignin = errors.New("You already signin")
)
