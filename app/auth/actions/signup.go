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

type SignupAction struct {
	// basic profile
	Name, Email string // required

	// Password and confirmation, required
	Password1 string
	Password2 string

	// Only valid when the role is: ROLE_INDIVIDUAL or ROLE_VENDOR, required
	Role model.UserRole

	// Optional fields, personal data
	Age                  uint8
	Address, City, State string

	// the result record if any
	user *model.User
}

func (action *SignupAction) Validate(cuser *model.User, service *service.Service) *validation.Result {
	if cuser != nil {
		return validation.Error(errors.New("Logout dulu untuk daftar akun"))
	}

	result := action.validateBasicInfo()
	if !result.Ok {
		return result
	}

	result = action.validatePassword()
	if !result.Ok {
		return result
	}

	userStorage, err := service.Storage.GetUserStorage()
	if err != nil {
		return validation.Error(err)
	}

	// make sure this email not used by other user
	_, err = userStorage.FindUserByEmail(context.Background(), action.Email)
	// if this return no nill then there are already user with this email
	if err == nil {
		result.AddFieldError("email", "email tersebut sudah dipakai, silakan login.")
		return result
	}

	// success
	user := &model.User{
		Profile: model.Profile{
			Name:    action.Name,
			Age:     action.Age,
			Address: action.Address,
			City:    action.City,
			State:   action.State,
		},
		Email: action.Email,
		Role:  action.Role,
		Credentials: model.Credentials{
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
		},
	}
	user.SetPassword([]byte(action.Password1))

	// insert
	objectid, err := userStorage.InsertUser(context.Background(), user)
	if err != nil {
		return validation.Error(err)
	}

	user.ID = objectid
	// store user
	action.user = user

	return result
}

func (action *SignupAction) GetUser() *model.User {
	return action.user
}

func (action *SignupAction) validateBasicInfo() *validation.Result {
	result := validation.Success()

	if action.Name == "" {
		result.AddFieldError("name", "Field nama harus diisi")
	}

	if action.Email == "" {
		result.AddFieldError("email", "email harus diisi")
	} else if !emailparser.IsValidEmail(action.Email) {
		result.AddFieldError("email", fmt.Sprintf("%s bukan email yang valid", action.Email))
	}

	return result
}

func (action *SignupAction) validatePassword() *validation.Result {
	result := validation.Success()

	length := len(action.Password1)
	if length < 8 {
		result.AddFieldError("password1", "panjang passwords minimal 8 karakter")
		return result
	}
	//not match?
	if action.Password1 != action.Password2 {
		result.AddFieldError("password2", "password dan konfirmasion password tidak sama")
	}

	return result
}
