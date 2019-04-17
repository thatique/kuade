package redistoken

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/thatique/kuade/app/model"
)

const (
	defaultRedisHost = "127.0.0.1"
	defaultRedisPort = "6379"
)

func dial(network, address string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return c, err
}

func createRedisPool() *redis.Pool {
	addr := os.Getenv("REDIS_HOST")
	if addr == "" {
		addr = defaultRedisHost
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = defaultRedisPort
	}

	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dial("tcp", fmt.Sprintf("%s:%s", addr, port))
		},
	}
}

func TestGenerateValid(t *testing.T) {
	gen := New(createRedisPool())
	user := &model.User{
		ID:    v1.NewObjectID(),
		Email: "foo@example.com",
		Credentials: model.Credentials{
			Enabled:    true,
			CreatedAt:  time.Now(),
			LastSignin: time.Now(),
		},
	}
	user.SetPassword([]byte("secret"))

	token, err := gen.Generate(user)
	if err != nil {
		t.Fatalf("Failed to generate password change token: %v", err)
	}
	if !gen.IsValid(user, token) {
		t.Fatalf("The generated password must be valid: %v", err)
	}
}

func TestGenerateValidLast(t *testing.T) {
	gen := New(createRedisPool())
	user := &model.User{
		ID:    v1.NewObjectID(),
		Email: "foo@example.com",
		Credentials: model.Credentials{
			Enabled:    true,
			CreatedAt:  time.Now(),
			LastSignin: time.Now(),
		},
	}
	user.SetPassword([]byte("secret"))
	token1, err := gen.Generate(user)
	if err != nil {
		t.Fatalf("Failed to generate password change token: %v", err)
	}
	token2, err := gen.Generate(user)
	if gen.IsValid(user, token1) {
		t.Fatalf("The old password token must be valid: %v", err)
	}
	if !gen.IsValid(user, token2) {
		t.Fatalf("The last password token must be valid: %v", err)
	}
}

func TestDeletedTokenShouldInvalid(t *testing.T) {
	gen := New(createRedisPool())
	user := &model.User{
		ID:    model.NewID(),
		Email: "foo@example.com",
		Credentials: model.Credentials{
			Enabled:    true,
			CreatedAt:  time.Now(),
			LastSignin: time.Now(),
		},
	}
	user.SetPassword([]byte("secret"))
	token, err := gen.Generate(user)

	// delete it
	err = gen.Delete(token)
	if err != nil {
		t.Fatalf("Failed to delete password change token: %v", err)
	}

	if gen.IsValid(user, token) {
		t.Fatalf("The deleted password change token should be invalid")
	}
}
