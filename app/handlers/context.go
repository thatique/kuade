package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
	"github.com/thatique/kuade/pkg/web/template"
)

const (
	plainContentType     = "text/plain"
	htmlContentType      = "text/html"
	jsonContentType      = "application/json"
	xmlContentType       = "application/xml"
	utf8PlainContentType = plainContentType + "; charset=utf-8"
	utf8HTMLContentType  = htmlContentType + "; charset=utf-8"
	utf8XMLContentType   = xmlContentType + "; charset=utf-8"
	utf8JSONContentType  = jsonContentType + "; charset=utf-8"
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

// Render render Golang HTML template using the context already in this object
// and extra.
func (ctx *Context) Render(w http.ResponseWriter, code int, extra template.M, tpls ...string) error {
	if ctx.tplContext != nil {
		for k, v := range ctx.tplContext {
			extra[k] = v
		}
	}

	return ctx.Blob(w, code, utf8HTMLContentType, ctx.htmlRenderer(extra, tpls...))
}

func (ctx *Context) htmlRenderer(extra template.M, tpls ...string) func(w io.Writer) error {
	return func(w io.Writer) error {
		ctx.renderer.Render(w, extra, tpls...)
		return nil
	}
}

// Blob write content type and content renderer
func (c *Context) Blob(w http.ResponseWriter, code int, contentType string, rd func(w io.Writer) error) error {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	return rd(w)
}

// Authenticate authenticate request
func (ctx *Context) Authenticate(r *http.Request) (*authenticator.Response, bool, error) {
	return ctx.authenticator.AuthenticateRequest(r)
}
