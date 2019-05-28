package cryptotoken

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"io"
	"testing"
	"time"

	"github.com/thatique/kuade/app/model"
)

func createKey() [32]byte {
	var key [32]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		panic(err)
	}
	return key
}

func TestMakeHashValue(t *testing.T) {
	now := time.Now().UTC()
	creds := &model.Credentials{
		Email:      "test@example.com",
		Username:   "test",
		UserID:     model.ID(1),
		Enabled:    true,
		CreatedAt:  now,
		LastSignin: now,
	}
	cred2 := &model.Credentials{
		Email:      "foo@example.com",
		Username:   "bar",
		UserID:     model.ID(2),
		Enabled:    true,
		CreatedAt:  now,
		LastSignin: now,
	}
	xs := makeHashValue(creds, numHours(time.Now().UTC()))
	ys := makeHashValue(cred2, numHours(time.Now().UTC()))
	if subtle.ConstantTimeCompare(xs, ys) == 1 {
		t.Fatal("expected different hash value returned")
	}
}

func TestCreateToken(t *testing.T) {
	key := createKey()
	now := time.Now().UTC()

	creds := &model.Credentials{
		Email:      "test@example.com",
		Username:   "test",
		UserID:     model.ID(1),
		Enabled:    true,
		CreatedAt:  now,
		LastSignin: now,
	}

	tokens := New(key, 2)
	token, err := tokens.Create(context.Background(), creds)
	if err != nil {
		t.Fatalf("failed to create token, %v", err)
	}

	if !tokens.Check(context.Background(), creds, token) {
		t.Error("expected token created to return true")
	}
}
