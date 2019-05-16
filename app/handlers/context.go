package handlers

import (
	"context"
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

func (ctx *Context) Plain(w http.ResponseWriter, code int, b []byte) error {
	w.Header().Set("Content-Type", utf8PlainContentType)
	w.WriteHeader(code)
	_, err := w.Write(b)
	return err
}

// Render render Golang HTML template using the context already in this object
// and extra.
func (ctx *Context) Render(w http.ResponseWriter, code int, tpls []string, extra template.M) error {
	if ctx.tplContext != nil {
		for k, v := range ctx.tplContext {
			extra[k] = v
		}
	}
	w.Header().Set("Content-Type", utf8HTMLContentType)
	w.WriteHeader(code)
	ctx.renderer.Render(w, extra, tpls...)
	return nil
}

// Authenticate authenticate request
func (ctx *Context) Authenticate(r *http.Request) (*authenticator.Response, bool, error) {
	return ctx.authenticator.AuthenticateRequest(r)
}
