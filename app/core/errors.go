package core

import "errors"

// ErrObjectDoesNotExist indicate the requested Object doesnot exist
var (
	ErrObjectDoesNotExist = errors.New("The requested object does not exist")

	// ErrImproperlyConfigured indicate our application not configured properly
	ErrImproperlyConfigured = errors.New("The application somehow improperly configured")
)
