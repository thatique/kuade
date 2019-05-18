package handlers

import (
	"net/http"
	"net/url"

	"github.com/thatique/kuade/app/auth/authenticator"
	"github.com/thatique/kuade/pkg/web/httputil"
)

// signputHandler logout user. The userID removed from session, and then redirect
// back to home or the redirect target defined in query string.
type signoutHandler struct {}

func (h *signoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := authenticator.Logout(r); err != nil {
		panic(err)
	}

	next := r.FormValue("next")
	if next == "" {
		next = "/"
	} else {
		next, _ = url.QueryUnescape(next)
	}

	if !httputil.IsSameSiteURLPath(next) {
		next = "/"
	}

	http.Redirect(w, r, next, http.StatusFound)
}
