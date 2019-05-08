package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
	"github.com/thatique/kuade/pkg/web/template"
)

// Context hold application and current request context
type Context struct {
	*App
	context.Context

	tplContext template.M
}

// Value overrides context.Context.Value to ensure that calls are routed to
// correct context.
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Context.Value(key)
}

// SetTplContext set template context
func (ctx *Context) SetTplContext(key string, val interface{}) {
	if ctx.tplContext == nil {
		ctx.tplContext = template.M{key: val}
	} else {
		ctx.tplContext[key] = val
	}
}

// RenderHTML render Golang HTML template using the context already in this object
// and extra.
func (ctx *Context) RenderHTML(w io.Writer, extra template.M, tpls ...string) {
	if ctx.tplContext != nil {
		for k, v := range ctx.tplContext {
			extra[k] = v
		}
	}
	ctx.renderer.Render(w, extra, tpls...)
}

// Authenticate authenticate request
func (ctx *Context) Authenticate(r *http.Request) (*authenticator.Response, bool, error) {
	return ctx.authenticator.AuthenticateRequest(r)
}
