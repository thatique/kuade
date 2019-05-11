package memory

import (
	"context"
	"flag"
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	"github.com/syaiful6/sersan"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/app/storage/driver"
)

func init() {
	storage.DefaultURLMux().RegisterStorage("memory", &urlopener{})
}

//
type urlopener struct{}

func (opener *urlopener) OpenStorageURL(ctx context.Context, ur *url.URL) (*storage.Store, error) {
	for param := range ur.Query() {
		return nil, fmt.Errorf("open storage %v: invalid query parameter %q", ur, param)
	}
	return storage.New(&memoryStorage{}), nil
}

// AddFlags adds CLI flags for configuring cassandra storage
func (opener *urlopener) AddFlags(flagSet *flag.FlagSet) {
}

// InitFromViper initializes configuring cassandra storage with properties from spf13/viper.
func (opener *urlopener) InitFromViper(v *viper.Viper) {
}

// OpenStorage open store backed with memory driver
func OpenStorage() *storage.Store {
	return storage.New(&memoryStorage{})
}

type memoryStorage struct {
	users    *userStore
	sessions *sessionStore
}

func (s *memoryStorage) GetUserStore() (driver.UserStore, error) {
	if s.users == nil {
		s.users = newUserStore()
	}

	return s.users, nil
}

func (s *memoryStorage) GetSessionStore() (sersan.Storage, error) {
	if s.sessions == nil {
		s.sessions = &sessionStore{sessions: make(map[string]*sersan.Session)}
	}

	return s.sessions, nil
}

func (s *memoryStorage) Close() error {
	return nil
}
