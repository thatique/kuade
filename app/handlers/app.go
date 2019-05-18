package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/syaiful6/sersan"

	auth "github.com/thatique/kuade/app/auth/authenticator"
	"github.com/thatique/kuade/app/storage"
	"github.com/thatique/kuade/pkg/iam/auth/authenticator"
	"github.com/thatique/kuade/pkg/iam/auth/request/union"
	"github.com/thatique/kuade/pkg/web/handlers"
	"github.com/thatique/kuade/pkg/web/template"
)

// App is our HTTP application
type App struct {
	context.Context

	asset         func(string) ([]byte, error)
	config        *Config
	renderer      *template.Renderer
	router        *mux.Router
	storage       *storage.Store
	authenticator authenticator.Request
}

// NewApp create Application
func NewApp(ctx context.Context, config *Config, asset func(string) ([]byte, error),
	storage *storage.Store) (app *App, err error) {

	users, err := storage.GetUserStorage()
	if err != nil {
		return nil, err
	}
	appAuth := union.New(auth.NewSessionAuthenticator(users))

	// router
	router := mux.NewRouter()
	router.StrictSlash(true)

	app = &App{
		Context:       ctx,
		asset:         asset,
		config:        config,
		storage:       storage,
		router:        router,
		authenticator: appAuth,
		renderer:      template.New(asset),
	}

	sessionState, err := app.configureSersan()
	if err != nil {
		return nil, err
	}

	webMiddlewares := handlers.NewIfRequestMiddleware([]mux.MiddlewareFunc{
		sersan.SessionMiddleware(sessionState),
		csrf.Protect(app.config.sessionKeys[0], csrf.Secure(app.config.httpSecure)),
	}, isNotAPIRoute)
	app.router.Use(webMiddlewares.Middleware)

	app.registerRoutes(users)

	return app, nil
}

// ServeHTTP is part of http.Handler interface
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}

// Dispatcher allow us to create http handler dynamically according to request
type Dispatcher interface {
	DispatchHTTP(ctx *Context, r *http.Request) http.Handler
}

// DispatcherFunc allow simple function to implements Dispatcher
type DispatcherFunc func(ctx *Context, r *http.Request) http.Handler

// DispatchHTTP is part of Dispatcher interface
func (d DispatcherFunc) DispatchHTTP(ctx *Context, r *http.Request) http.Handler {
	return d(ctx, r)
}

func (app *App) dispatchFunc(dispatch DispatcherFunc) http.Handler {
	return app.dispatch(dispatch)
}

func (app *App) dispatch(dsp Dispatcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			App:     app,
			Context: r.Context(),
		}
		r = r.WithContext(ctx)
		dsp.DispatchHTTP(ctx, r).ServeHTTP(w, r)
	})
}

func (app *App) registerRoutes(users *storage.UserStore) {
	app.router.NotFoundHandler = &pageHandler{
		Context: &Context{
			App: app,
		},
		title:       "Not Found",
		description: "The page you are looking for is not found",
		templates:   []string{"base.html", "404.html"},
		status:      http.StatusNotFound,
	}
	app.router.Handle("/", app.dispatch(&pageDispatcher{
		status:      http.StatusOK,
		title:       "Thatique - homepage",
		description: "Thatique description",
		templates:   []string{"base.html", "homepage.html"}})).Name("homepage")
	// static files
	app.router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(handlers.NewStaticFS("assets/static", app.asset))))

	// auth
	authRouter := app.router.PathPrefix("/auth").Subrouter()
	authRouter.Handle("/", app.dispatch(
		newSigninDispatcher(auth.NewPasswordAuthenticator(users), 3, 5))).Name("auth.signin")
	authRouter.Handle("/logout", &signoutHandler{}).Name("auth.signout")
}

func (app *App) configureSersan() (*sersan.ServerSessionState, error) {
	sessionStore, err := app.storage.GetSessionStore()
	if err != nil {
		return nil, err
	}
	if len(app.config.sessionKeys) == 0 {
		return nil, errors.New("http session keys must not be empty")
	}

	sessionstate := sersan.NewServerSessionState(sessionStore, app.config.sessionKeys...)
	sessionstate.AuthKey = auth.UserSessionKey
	sessionstate.Options.Secure = app.config.httpSecure

	return sessionstate, nil
}

func isNotAPIRoute(r *http.Request) bool {
	return !strings.HasPrefix(r.URL.Path, "/api")
}
