package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/kuade/service"
	"github.com/thatique/kuade/kuade/validation"
	"github.com/thatique/kuade/pkg/emailparser"
)

type SignupAction struct {
	// basic profile
	Name, Email string // required

	// Password and confirmation, required
	Password1  string
	Password2  string

	// Only valid when the role is: ROLE_INDIVIDUAL or ROLE_VENDOR, required
	Role       auth.Role

	// Optional fields, personal data
	Age        uint8
	Address, City, State string

	// the result record if any
	user *auth.User
}

func (action *SignupAction) Validate(cuser *auth.User, service *service.Service) *validation.Result {
	if cuser != nil && cuser.Role != auth.ROLE_STAFF && cuser.Role != auth.ROLE_SUPERUSER {
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

	userStore, err := service.Storage.GetUserStorage()
	if err != nil {
		return validation.Error(err)
	}

	// make sure this email not used by other user
	_, err = userStore.FindByEmail(context.Background(), action.Email)
	// if this return no nill then there are already user with this email
	if err == nil {
		result.AddFieldError("email", "email tersebut sudah dipakai, silakan login.")
		return result
	}

	if action.Role != auth.ROLE_INDIVIDUAL && action.Role != auth.ROLE_VENDOR {
		result.AddFieldError("role", "Silakan pilih ingin menjadi akun vendor atau individual")
		return result
	}

	// success
	user := &auth.User{
		Profile: auth.Profile{
			Name: action.Name,
			Age:  action.Age,
			Address: action.Address,
			City: action.City,
			State: action.State,
		},
		Email: action.Email,
		Role:  action.Role,
		Credentials: auth.Credentials{
			Enabled: true,
			CreatedAt: time.Now().UTC(),
		},
	}
	user.SetPassword([]byte(action.Password1))

	// insert
	objectid, err := userStore.Create(context.Background(), user)
	if err != nil {
		return validation.Error(err)
	}

	user.Id = objectid
	// store user
	action.user = user

	return result
}

func (action *SignupAction) GetUser() *auth.User {
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
