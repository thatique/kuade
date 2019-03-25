package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/service"
	"github.com/thatique/kuade/kuade/validation"
	"github.com/thatique/kuade/pkg/emailparser"
)

// SignInAction represent user action input when signin by `Email` and `Password`
type SignInAction struct {
	Email    string
	Password string

	user     *auth.User
}

func (action *SignInAction) Validate(user *auth.User, service *service.Service) *validation.Result {
	if user != nil {
		return validation.Error(auth.ErrAlreadySignin)
	}

	result := action.validateInput()
	if !result.Ok {
		action.runHasherIfFailed()
		return result
	}

	userStore, err := service.Storage.GetUserStorage()
	if err != nil {
		return validation.Error(err)
	}

	user, err = userStore.FindByEmail(context.Background(), action.Email)
	if err != nil {
		action.runHasherIfFailed()
		return validation.Error(err)
	}

	if !user.IsActive() {
		var msg string
		if user.Status == auth.USER_STATUS_INACTIVE {
			msg = "status akun anda tidak aktif"
		} else {
			msg = "status akun anda sedang terkunci"
		}

		result.AddFieldError("status", msg)
		action.runHasherIfFailed()
		return result
	}

	if !user.VerifyPassword([]byte(action.Password)) {
		result.AddFieldError("email_password", "email atau password tersebut salah.")
		return result
	}

	credentials := user.Credentials
	credentials.LastSignin = time.Now().UTC()

	//update last login
	err = userStore.UpdateCredentials(context.Background(), user.Id, credentials)
	if err != nil {
		return validation.Error(err)
	}

	user.Credentials = credentials
	action.user = user

	return validation.Success()
}

func (action *SignInAction) GetUser() *auth.User {
	return action.user
}

func (action *SignInAction) runHasherIfFailed() {
	var user *auth.User
	user.SetPassword([]byte(action.Password))
}

func (action *SignInAction) validateInput() (result *validation.Result) {
	result = validation.Success()

	if action.Email == "" {
		result.AddFieldError("email", "email harus diisi")
	} else if !emailparser.IsValidEmail(action.Email) {
		result.AddFieldError("email", fmt.Sprintf("%s bukan email yang valid", action.Email))
	}

	if action.Password == "" {
		result.AddFieldError("password", "password harus diisi")
	}

	return
}
