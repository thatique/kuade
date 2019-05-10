package context

import "context"

// key for version
type versionKey struct{}

// WithVersion stores the application version in the context.
func WithVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionKey{}, version)
}

func GetVersion(ctx context.Context) (version string) {
	if value, ok := ctx.Value(versionKey{}).(string); ok {
		version = value
	}
	return version
}
