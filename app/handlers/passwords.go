package handlers

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/thatique/kuade/app/auth/passwords"
)

const (
	userPasswordPath        = "/users/passwords"
	internalSetToken        = "_set-password"
	internalSetTokenSession = "_password_token_"
)

type passwordsDispatcher struct {
	broker *passwords.Broker
}

type passwordsResetHandler struct {
	*Context

	broker *passwords.Broker
}

func (p *passwordsDispatcher) dispatchPasswordReset(ctx *Context, r *http.Request) http.Handler {
	hd := &passwordsDispatcher{
		Context: ctx,
		broker: p.broker,
	}

	return hd
}

func (h *passwordsResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
