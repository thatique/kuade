package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	gorHandler "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/syaiful6/sersan"
	"github.com/thatique/kuade/app/auth/authenticator"
	"github.com/thatique/kuade/app/auth/passwords"
	"github.com/thatique/kuade/pkg/web/handlers"
	"github.com/thatique/kuade/pkg/web/httputil"
	"github.com/thatique/kuade/pkg/web/template"
)

const (
	userPasswordPath        = "/users/passwords"
	internalSetToken        = "_set-password"
	internalSetTokenSession = "_password_token_"
)

type passwordsDispatcher struct {
	broker *passwords.Broker
	limiter *handlers.RateLimiter
}

type passwordsHandler struct {
	*Context
	broker *passwords.Broker
	limiter *handlers.RateLimiter
	request *passwords.Request
}

func (p *passwordsDispatcher) dispatchResetLink(ctx *Context, r *http.Request) http.Handler {
	h := passwordsHandler{
		Context: ctx,
		broker: p.broker,
		limiter: p.limiter,
	}

	return gorHandler.MethodHandler{
		"GET": http.HandlerFunc(h.renderResetLink),
		"POST": http.HandlerFunc(h.sendResetLink),
	}
}

func (p *passwordsDispatcher) dispatchResetPassword(ctx *Context, r *http.Request) http.Handler {
	vars := mux.Vars(r)
	var (
		uid   = vars["uid"]
		token = vars["token"]
	)

	sess, err := sersan.GetSession(r)
	if err != nil {
		panic(err)
	}
	if token == internalSetToken {
		if v, ok := sess[internalSetTokenSession]; ok {
			token = v.(string)
		}

		req, ok := p.broker.ValidateReset(r.Context(), vars["uid"], token)
		if !ok {
			// it's not valid request
			return pageHandler{
				Context: ctx,
				status: http.StatusForbidden,
				title: "403 Forbidden",
				Description: "Invalid URL",
				templates: []string{"base.html", "403.html"},
			}
		}
		h := passwordsHandler{
			Context: ctx,
			broker: p.broker,
			limiter: p.limiter,
			request: req,
		}
		return gorHandler.MethodHandler{
			"GET": http.HandlerFunc(h.renderChangePassword),
			"POST": http.HandlerFunc(h.changePassword),
		}
	}

	sess[internalSetTokenSession] = token
	url := fmt.Sprintf("/auth/passwords/%s/%s", uid, internalSetToken)
	return http.RedirectHandler(url, http.StatusFound)
}

func (h *passwordsHandler) templateContext(link bool, r *http.Request) {
	h.SetTplContext(csrf.TemplateTag, csrf.TemplateField(r))
	if link {
		h.SetTplContext("Title", "Reset Password - Thatique")
		h.SetTplContext("Description", "Reset your Password")
		return
	}
	h.SetTplContext("Title", "Change Password - Thatique")
	h.SetTplContext("Description", "Change your Password")
}

func (h *passwordsHandler) renderResetLink(w http.ResponseWriter, r *http.Request) {
	h.templateContext(true, r)
	h.Render(w, http.StatusOK, []string{"base.html", "auth/password_reset_form.html"}, template.M{"Email": ""})
}

func (h *passwordsHandler) renderChangePassword(w http.ResponseWriter, r *http.Request) {
	h.templateContext(false, r)
	h.Render(w, http.StatusOK, []string{"base.html", "auth/password_change_form.html"}, template.M{})
}

func (h *passwordsHandler) sendResetLink(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.Plain(w, http.StatusBadRequest, []byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	h.templateContext(true, r)
	limiter := h.limiter.Get(r)
	if !limiter.Allow() {
		h.Render(w, http.StatusTooManyRequests, []string{"base.html", "auth/password_reset_form.html"}, template.M{
			"Errors": []string{http.StatusText(429)},
			"Email": "",
		})
		return
	}

	err = h.broker.SendResetLink(r.Context(), httputil.GetSourceIP(r), r.FormValue("email"))
	if err != nil {
		h.Render(w, http.StatusBadRequest, []string{"base.html", "auth/password_reset_form.html"}, template.M{
			"Errors": []string{err.Error()},
			"Email": r.FormValue("email"),
		})
		return
	}

	// success
	h.Render(w, http.StatusOK, []string{"base.html", "auth/password_reset_form.html"}, template.M{
		"Errors": []string{"Password reset sudah terkirim ke inbox anda"}
		"Email": "",
	})
}

func (h *passwordsHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.Plain(w, http.StatusBadRequest, []byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	p.request.Password1 = r.FormValue("password")
	p.request.Password2 = r.FormValue("confirm-password")
	if errCode := r.broker.Resets(r.Context(), p.request); errCode != passwords.NoError {
		h.templateContext(false, r)
		h.Render(w, http.StatusBadRequest, []string{"base.html", "auth/password_reset_form.html"}, template.M{
			"Errors": []string{errCode.ErrorDescription()}
		})
		return
	}
	sess, err := sersan.GetSession(r)
	if err != nil {
		panic(err)
	}
	delete(sess, internalSetTokenSession)

	// let it login then redirect
	authenticator.Login(r, p.request.GetUser())
	http.Redirect(w, r, "/", http.StatusFound)
}
