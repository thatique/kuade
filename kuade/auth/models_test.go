package auth

import (
	"testing"
)

func TestPasswordHasher(t *testing.T) {
	userCases := []struct{
		email, password, checkpassword string
		correct bool
	}{
		{
			email: "nami@pub.example.com",
			password: "secret",
			checkpassword: "secret",
			correct: true,
		},
		{
			email: "nami@pub.example.com",
			password: "secret",
			checkpassword: "secret2",
			correct: false,
		},
		{
			email:    "luci@machine.example",
			password: "secret12333longpasswordssssssssssssssaaaaaa",
			checkpassword: "secret12333longpasswordssssssssssssssaaaaaa",
			correct: true,
		},
	}

	for i, data := range userCases {
		user := &User{Email: data.email}
		err := user.SetPassword([]byte(data.password))
		if err != nil {
			t.Fatalf("Failed to create user for %d", i)
			return
		}

		if user.VerifyPassword(data.checkpassword) != data.correct {
			t.Errorf("passwords for %s should return %v", user.Email, data.correct)
			return
		}
	}
}