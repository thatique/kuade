package factory

import (
	"fmt"

	"github.com/thatique/kuade/app/storage"
)

// driverFactories stores an internal mapping between storage driver names and their respective
// factories
var driverFactories = make(map[string]StorageDriverFactory)

type StorageDriverFactory interface {
	// Create returns a new storagedriver.StorageDriver with the given parameters
	// Parameters will vary by driver and may be ignored
	// Each parameter key must only consist of lowercase letters and numbers
	Create(parameters map[string]interface{}) (storage.Driver, error)
}

type StorageDriverFactoryFn func(parameters map[string]interface{}) (storage.Driver, error)

func (fc StorageDriverFactoryFn) Create(parameters map[string]interface{}) (storage.Driver, error) {
	return fc(parameters)
}

// Register makes a storage driver available by the provided name.
// If Register is called twice with the same name or if driver factory is nil, it panics.
// Additionally, it is not concurrency safe. Most Storage Drivers call this function
// in their init() functions. See the documentation for StorageDriverFactory for more.
func Register(name string, factory StorageDriverFactory) {
	if factory == nil {
		panic("Must not provide nil StorageDriverFactory")
	}
	_, registered := driverFactories[name]
	if registered {
		panic(fmt.Sprintf("StorageDriverFactory named %s already registered", name))
	}

	driverFactories[name] = factory
}

// Register using function as factory
func RegisterFunc(name string, factory StorageDriverFactoryFn) {
	Register(name, factory)
}

// Create a new storagedriver.StorageDriver with the given name and
// parameters. To use a driver, the StorageDriverFactory must first be
// registered with the given name. If no drivers are found, an
// InvalidStorageDriverError is returned
func Create(name string, parameters map[string]interface{}) (storage.Driver, error) {
	driverFactory, ok := driverFactories[name]
	if !ok {
		return nil, InvalidStorageDriverError{name}
	}
	return driverFactory.Create(parameters)
}

// InvalidStorageDriverError records an attempt to construct an unregistered storage driver
type InvalidStorageDriverError struct {
	Name string
}

func (err InvalidStorageDriverError) Error() string {
	return fmt.Sprintf("StorageDriver not registered: %s", err.Name)
}
