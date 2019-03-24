package handlers

import (
	"context"

	"github.com/thatique/kuade/kuade/auth"
	webcontext "github.com/thatique/kuade/pkg/web/context"
)

type Context struct {
	*App
	context.Context
}

// Value overrides context.Context.Value to ensure that calls are routed to
// correct context.
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Context.Value(key)
}

func (ctx *Context) User() (*auth.User, error) {
	r, err := webcontext.GetRequest(ctx)
	if err != nil {
		return nil, err
	}
	user := ctx.authSess.User(r)
	return user, nil
}

func getName(ctx context.Context) (name string) {
	return webcontext.GetStringValue(ctx, "vars.name")
}
