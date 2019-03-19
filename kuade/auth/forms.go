package auth

import (
	"context"
	"fmt"

	"github.com/thatique/kuade/pkg/emailparser"
)

// signinform represent form input that user submit when trying to login
type SigninForm struct {
	users UserStore

	Email    string
	Password string
}

func (form *SigninForm) Validate() (user *User, err map[string]string, ok bool) {
	err = make(map[string]string)
	ok = form.validateInput(err)
	if !ok {
		return
	}

	user, verr := form.users.FindByEmail(context.Background(), form.Email)
	if verr != nil {
		ok = false
		err["email_password"] = "email atau password Anda keliru"
		return
	}

	if !user.IsActive() {
		var msg string
		if user.Status == USER_STATUS_INACTIVE {
			msg = "status akun anda tidak aktif"
		} else {
			msg = "status akun anda sedang terkunci"
		}

		err["status"] = msg
		ok = false
		return
	}

	if !user.VerifyPassword([]byte(form.Password)) {
		ok = false
		err["email_password"] = "email atau password anda keliru"
		return
	}

	ok = true
	return
}

func (form *SigninForm) validateInput(m map[string]string) bool {
	ok := true
	if form.Email == "" {
		m["Email"] = "field email harus diisi"
		ok = false
	} else if !emailparser.IsValidEmail(form.Email) {
		m["Email"] = fmt.Sprintf("%s bukan sebuah email", form.Email)
		ok = false
	}

	if form.Password == "" {
		m["Password"] = "password harus diisi"
		ok = false
	}

	return ok
}
