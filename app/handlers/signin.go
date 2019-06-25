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

type signinDispatcher struct {
	auth    authenticator.Password
	limiter *handlers.RateLimiter
}

type signinHandler struct {
	*Context
	auth    authenticator.Password
	limiter *handlers.RateLimiter
}

func newSigninDispatcher(auth authenticator.Password, n, b int) *signinDispatcher {
	return &signinDispatcher{
		auth:    auth,
		limiter: handlers.NewRateLimiter(n, b, httputil.GetSourceIP),
	}
}

func (d *signinDispatcher) DispatchHTTP(ctx *Context, r *http.Request) http.Handler {
	if _, ok, err := ctx.Authenticate(r); ok && err == nil {
		return http.RedirectHandler("/", http.StatusFound)
	}

	var sh = &signinHandler{Context: ctx, auth: d.auth, limiter: d.limiter}

	return gorHandler.MethodHandler{
		"GET":  http.HandlerFunc(sh.showSigninForm),
		"POST": http.HandlerFunc(sh.postSigninForm),
	}
}

func (h *signinHandler) templateContext(r *http.Request) {
	h.SetTplContext("Title", "Signin - Thatique")
	h.SetTplContext("Description", "Signin to Thatique")
	h.SetTplContext(csrf.TemplateTag, csrf.TemplateField(r))
}

func (h *signinHandler) showSigninForm(w http.ResponseWriter, r *http.Request) {
	h.templateContext(r)
	h.Render(w, http.StatusOK, []string{"base.html", "auth/signin.html"}, template.M{})
}

func (h *signinHandler) postSigninForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// this is likely parsing error
		h.Plain(w, http.StatusBadRequest, []byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	limiter := h.limiter.Get(r)
	if !limiter.Allow() {
		h.templateContext(r)
		h.SetTplContext("Errors", []string{http.StatusText(429)})
		h.Render(w, http.StatusTooManyRequests, []string{"base.html", "auth/signin.html"}, template.M{})
		return
	}

	response, ok, err := h.auth.AuthenticatePassword(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if !ok || err != nil {
		h.templateContext(r)
		h.SetTplContext("Errors", []string{"Password atau username yang Anda masukkan keliru"})
		h.Render(w, http.StatusUnauthorized, []string{"base.html", "auth/signin.html"}, template.M{})
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
