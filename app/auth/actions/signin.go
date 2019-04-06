package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/app/service"
	"github.com/thatique/kuade/app/validation"
	"github.com/thatique/kuade/pkg/emailparser"
)

// SignInAction represent user action input when signin by `Email` and `Password`
type SignInAction struct {
	Email    string
	Password string

	user *model.User
}

func (action *SignInAction) Validate(user *model.User, service *service.Service) *validation.Result {
	if user != nil {
		return validation.Error(errors.New("Kamu sudah login"))
	}

	result := action.validateInput()
	if !result.Ok {
		action.runHasherIfFailed()
		return result
	}

	userStorage, err := service.Storage.GetUserStorage()
	if err != nil {
		return validation.Error(err)
	}

	user, err = userStorage.FindUserByEmail(context.Background(), action.Email)
	if err != nil {
		action.runHasherIfFailed()
		return validation.Error(err)
	}

	if !user.IsActive() {
		var msg string
		if user.Status == model.UserStatus_INACTIVE {
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
	err = userStorage.UpdateUserCredentials(context.Background(), user.ID, credentials)
	if err != nil {
		return validation.Error(err)
	}

	user.Credentials = credentials
	action.user = user

	return validation.Success()
}

func (action *SignInAction) GetUser() *model.User {
	return action.user
}

func (action *SignInAction) runHasherIfFailed() {
	var user *model.User
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
