package handlers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"
	gorHandler "github.com/gorilla/handlers"
	appAuth "github.com/thatique/kuade/app/auth/authenticator"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
	"github.com/thatique/kuade/pkg/web/handlers"
	"github.com/thatique/kuade/pkg/web/httputil"
	"github.com/thatique/kuade/pkg/web/template"
)

type signinHandler struct {
	*Context
	auth    authenticator.Password
	limiter *handlers.Ratelimiter
}

func siginDispatcher(auth authenticator.Password, limiter *handlers.RateLimiter) func(ctx *Context, r *http.Request) http.Handler {
	return func(ctx *Context, r *http.Request) http.Handler {
		// if user already logged in, redirect to homepage
		if _, ok, err := ctx.Authenticate(r); ok && err == nil {
			return http.RedirectHandler("/", http.StatusFound)
		}
		var sh = &siginHandler{Context: ctx, auth: auth, limiter: limiter}

		return gorHandler.MethodHandler{
			"GET":  http.HandleFunc(sh.showSigninForm),
			"POST": http.HandleFunc(sh.postSigninForm),
		}
	}
}

func (h *signinHandler) templateContext(r *http.Request) {
	h.SetTplContext("Title", "Signin - Thatique")
	h.SetTplContext("Description", "Signin to Thatique")
	h.SetTplContext(csrf.TemplateTag, csrf.TemplateField(r))
}

func (h *signinHandler) showSigninForm(w http.ResponseWriter, r *http.Request) {
	h.templateContext(r)
	h.Render(w, http.StatusOK, template.M{}, []string{"base.html", "auth/signin.html"})
}

func (h *signinHandler) postSigninForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// this is likely parsing error
		h.Plain(w, http.StatusBadRequest, []byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	limiter := h.limiter.Get(r)
	if limiter.Allow() == false {
		h.templateContext(r)
		h.SetTplContext("Errors", []string{http.StatusText(429)})
		h.Render(w, http.StatusTooManyRequests, template.M, []string{"base.html", "auth/signin.html"})
		return
	}

	response, ok, err := h.auth.AuthenticatePassword(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if !ok || err != nil {
		h.templateContext(r)
		h.SetTplContext("Errors", []string{"Password atau username yang Anda masukkan keliru"})
		h.Render(w, http.StatusUnauthorized, template.M{}, []string{"base.html", "auth/signin.html"})
		return
	}

	// TODO: handler error
	user := response.User.(*model.User)
	appAuth.Login(r, user)

	redirectURL := r.FormValue("next")
	if redirectURL == "" {
		redirectURL = "/"
	} else {
		redirectURL, _ = url.QueryUnescape(redirectURL)
	}

	if !httputil.IsSameSiteURLPath(redirectURL) {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
