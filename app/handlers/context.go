package handlers

import (
	"context"

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

func getName(ctx context.Context) (name string) {
	return webcontext.GetStringValue(ctx, "vars.name")
}
