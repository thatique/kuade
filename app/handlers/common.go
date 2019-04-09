package handlers

import (
	"net/http"
	"github.com/gorilla/handlers"
	"github.com/thatique/kuade/pkg/web/template"
)

type pageHandler struct {
	*Context

	title, description string

	templates []string
}

func (p *pageHandler) DispatchHTTP(ctx *Context, r *http.Request) http.Handler {
	p.Context = ctx

	ctx.SetTplContext("Title", p.title)
	ctx.SetTplContext("Description", p.description)

	return handlers.MethodHandler{
		"GET":     p,
		"OPTIONS": p,
	}
}

func (p *pageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response, ok, _ := p.Authenticate(r)
	if ok {
		p.SetTplContext("User", response.User)
	}
	p.RenderHTML(w, template.M{}, p.templates...)
}
