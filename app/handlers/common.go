package handlers

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/thatique/kuade/pkg/web/template"
)

type pageDispatcher struct {
	status int

	title, description string

	templates []string
}

type pageHandler struct {
	*Context
	// http status code
	status int

	title, description string

	templates []string
}

func (d *pageDispatcher) DispatchHTTP(ctx *Context, r *http.Request) http.Handler {
	var h = &pageHandler{
		Context:     ctx,
		status:      d.status,
		title:       d.title,
		description: d.description,
		templates:   d.templates,
	}

	return handlers.MethodHandler{
		"GET":     h,
		"OPTIONS": h,
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
	if err := p.Render(w, statusCode, p.templates, template.M{}); err != nil {
		panic(err)
	}
}
