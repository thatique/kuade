package handlers

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/thatique/kuade/pkg/web/template"
)

type pageHandler struct {
	*Context
	// http status code
	status int

	title, description string

	templates []string
}

func (p *pageHandler) DispatchHTTP(ctx *Context, r *http.Request) http.Handler {
	p.Context = ctx

	return handlers.MethodHandler{
		"GET":     p,
		"OPTIONS": p,
	}
}

func (p *pageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.SetTplContext("Title", p.title)
	p.SetTplContext("Description", p.description)
	response, ok, _ := p.Authenticate(r)
	if ok {
		p.SetTplContext("User", response.User)
	}
	var statusCode = http.StatusOK
	if p.status != 0 {
		statusCode = p.status
	}
	if err := p.Render(w, statusCode, template.M{}, p.templates...); err != nil {
		panic(err)
	}
}
