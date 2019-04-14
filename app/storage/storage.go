package storage

import (
	"context"
	"net/url"

	"github.com/thatique/kuade/app/storage/driver"
	"github.com/thatique/kuade/pkg/metrics"
	"github.com/thatique/kuade/pkg/openurl"
)

const pkgName = "github.com/thatique/kuade/app/storage"

var (
	latencyMeasure = metrics.LatencyMeasure(pkgName)

	// OpenCensusViews are predefined views for OpenCensus metrics.
	// The views include counts and latency distributions for API method calls.
	// See the example at https://godoc.org/go.opencensus.io/stats/view for usage.
	OpenCensusViews = metrics.Views(pkgName, latencyMeasure)
)

// Storage wrapped storage driver with useful metric
type Storage struct {
	driver driver.Driver

	user *UserStorage
}

// New create storage
func New(driver driver.Driver) *Storage {
	return &Storage{driver: driver}
}

// GetUserStorage get wrapped user storage
func (s *Storage) GetUserStorage() (*UserStorage, error) {
	if s.user != nil {
		return s.user, nil
	}
	us, err := s.driver.GetUserStorage()
	if err != nil {
		return nil, err
	}
	s.user = NewUserStorage(us)
	return s.user, nil
}

// URLOpener can create Storage based URL
type URLOpener interface {
	OpenStorageURL(ctx context.Context, u *url.URL) (*Storage, error)
}

// URLMux is store registered storage driver
type URLMux struct {
	schemes openurl.SchemaMap
}

// RegisterStorage register storage to URLMux
func (mux *URLMux) RegisterStorage(scheme string, opener URLOpener) {
	mux.schemes.Register("storage", "Storage", scheme, opener)
}

// OpenStorage open storage based URL string
func (mux *URLMux) OpenStorage(ctx context.Context, urlstr string) (*Storage, error) {
	opener, u, err := mux.schemes.FromString("Storage", urlstr)
	if err != nil {
		return nil, err
	}
	return opener.(URLOpener).OpenTransportURL(ctx, u)
}

// OpenStorageURL dispatches the URL to the opener that is registered with the
// URL's scheme. OpenTransportURL is safe to call from multiple goroutines.
func (mux *URLMux) OpenStorageURL(ctx context.Context, u *url.URL) (*Storage, error) {
	opener, err := mux.schemes.FromURL("Storage", u)
	if err != nil {
		return nil, err
	}
	return opener.(URLOpener).OpenTransportURL(ctx, u)
}

var defaultURLMux = new(URLMux)

// DefaultURLMux returns the URLMux used by OpenStorage.
//
// Driver packages can use this to register their URLOpener on the mux.
func DefaultURLMux() *URLMux {
	return defaultURLMux
}

// OpenStorage opens the Keeper identified by the URL given.
// See the URLOpener documentation in provider-specific subpackages for
// details on supported URL formats
func OpenStorage(ctx context.Context, urlstr string) (*Transport, error) {
	return defaultURLMux.OpenStorage(ctx, urlstr)
}
